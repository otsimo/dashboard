package dashboard

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	phttp "github.com/coreos/go-oidc/http"
	"github.com/coreos/go-oidc/jose"
	"github.com/coreos/go-oidc/key"
	"github.com/coreos/go-oidc/oauth2"
	"github.com/coreos/go-oidc/oidc"
	"github.com/coreos/pkg/timeutil"
	"github.com/jonboulle/clockwork"
)

const (
	// amount of time that must pass after the last key sync
	// completes before another attempt may begin
	keySyncWindow               = 5 * time.Second
	MaximumTokenRefreshInterval = 6 * time.Hour
	MinimumTokenRefreshInterval = time.Minute
)

var (
	supportedAuthMethods = map[string]struct{}{
		oauth2.AuthMethodClientSecretBasic: struct{}{},
		oauth2.AuthMethodClientSecretPost:  struct{}{},
	}
)

func NewOIDCClient(cfg oidc.ClientConfig) (*Client, error) {
	// Allow empty redirect URL in the case where the client
	// only needs to verify a given token.
	ru, err := url.Parse(cfg.RedirectURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redirect URL: %v", err)
	}

	c := Client{
		credentials:    cfg.Credentials,
		httpClient:     cfg.HTTPClient,
		scope:          cfg.Scope,
		redirectURL:    ru.String(),
		providerConfig: newProviderConfigRepo(cfg.ProviderConfig),
		keySet:         cfg.KeySet,
	}

	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	if c.scope == nil {
		c.scope = make([]string, len(oidc.DefaultScope))
		copy(c.scope, oidc.DefaultScope)
	}
	return &c, nil
}

type Client struct {
	httpClient     phttp.Client
	providerConfig *providerConfigRepo
	credentials    oidc.ClientCredentials
	redirectURL    string
	scope          []string
	keySet         key.PublicKeySet
	providerSyncer *oidc.ProviderConfigSyncer

	keySetSyncMutex sync.RWMutex
	lastKeySetSync  time.Time
}

func (c *Client) Healthy() error {
	now := time.Now().UTC()

	cfg := c.providerConfig.Get()

	if cfg.Empty() {
		return errors.New("oidc client provider config empty")
	}

	if !cfg.ExpiresAt.IsZero() && cfg.ExpiresAt.Before(now) {
		return errors.New("oidc client provider config expired")
	}

	return nil
}

func (c *Client) OAuthClient() (*oauth2.Client, error) {
	cfg := c.providerConfig.Get()
	authMethod, err := chooseAuthMethod(cfg)
	if err != nil {
		return nil, err
	}

	ocfg := oauth2.Config{
		Credentials: oauth2.ClientCredentials(c.credentials),
		RedirectURL: c.redirectURL,
		AuthURL:     cfg.AuthEndpoint.String(),
		TokenURL:    cfg.TokenEndpoint.String(),
		Scope:       c.scope,
		AuthMethod:  authMethod,
	}

	return oauth2.NewClient(c.httpClient, ocfg)
}

func chooseAuthMethod(cfg oidc.ProviderConfig) (string, error) {
	if len(cfg.TokenEndpointAuthMethodsSupported) == 0 {
		return oauth2.AuthMethodClientSecretBasic, nil
	}

	for _, authMethod := range cfg.TokenEndpointAuthMethodsSupported {
		if _, ok := supportedAuthMethods[authMethod]; ok {
			return authMethod, nil
		}
	}

	return "", errors.New("no supported auth methods")
}

// SyncProviderConfig starts the provider config syncer
func (c *Client) SyncProviderConfig(discoveryURL string) chan struct{} {
	r := oidc.NewHTTPProviderConfigGetter(c.httpClient, discoveryURL)
	s := oidc.NewProviderConfigSyncer(r, c.providerConfig)
	stop := s.Run()
	s.WaitUntilInitialSync()
	return stop
}

func (c *Client) maybeSyncKeys() error {
	tooSoon := func() bool {
		return time.Now().UTC().Before(c.lastKeySetSync.Add(keySyncWindow))
	}

	// ignore request to sync keys if a sync operation has been
	// attempted too recently
	if tooSoon() {
		return nil
	}

	c.keySetSyncMutex.Lock()
	defer c.keySetSyncMutex.Unlock()

	// check again, as another goroutine may have been holding
	// the lock while updating the keys
	if tooSoon() {
		return nil
	}

	cfg := c.providerConfig.Get()
	r := oidc.NewRemotePublicKeyRepo(c.httpClient, cfg.KeysEndpoint.String())
	w := &clientKeyRepo{client: c}
	_, err := key.Sync(r, w)
	c.lastKeySetSync = time.Now().UTC()

	return err
}

type clientKeyRepo struct {
	client *Client
}

func (r *clientKeyRepo) Set(ks key.KeySet) error {
	pks, ok := ks.(*key.PublicKeySet)
	if !ok {
		return errors.New("unable to cast to PublicKey")
	}
	r.client.keySet = *pks
	return nil
}

func (c *Client) ClientCredsToken(scope []string) (jose.JWT, error) {
	cfg := c.providerConfig.Get()

	if !cfg.SupportsGrantType(oauth2.GrantTypeClientCreds) {
		return jose.JWT{}, fmt.Errorf("%v grant type is not supported", oauth2.GrantTypeClientCreds)
	}

	oac, err := c.OAuthClient()
	if err != nil {
		return jose.JWT{}, err
	}

	t, err := oac.ClientCredsToken(scope)
	if err != nil {
		return jose.JWT{}, err
	}

	jwt, err := jose.ParseJWT(t.IDToken)
	if err != nil {
		return jose.JWT{}, err
	}

	return jwt, c.VerifyJWT(jwt)
}

// ExchangeAuthCode exchanges an OAuth2 auth code for an OIDC JWT ID token.
func (c *Client) ExchangeAuthCode(code string) (jose.JWT, error) {
	oac, err := c.OAuthClient()
	if err != nil {
		return jose.JWT{}, err
	}

	t, err := oac.RequestToken(oauth2.GrantTypeAuthCode, code)
	if err != nil {
		return jose.JWT{}, err
	}

	jwt, err := jose.ParseJWT(t.IDToken)
	if err != nil {
		return jose.JWT{}, err
	}

	return jwt, c.VerifyJWT(jwt)
}

// RefreshToken uses a refresh token to exchange for a new OIDC JWT ID Token.
func (c *Client) RefreshToken(refreshToken string) (jose.JWT, error) {
	oac, err := c.OAuthClient()
	if err != nil {
		return jose.JWT{}, err
	}

	t, err := oac.RequestToken(oauth2.GrantTypeRefreshToken, refreshToken)
	if err != nil {
		return jose.JWT{}, err
	}

	jwt, err := jose.ParseJWT(t.IDToken)
	if err != nil {
		return jose.JWT{}, err
	}

	return jwt, c.VerifyJWT(jwt)
}

func (c *Client) VerifyJWT(jwt jose.JWT) error {
	var keysFunc func() []key.PublicKey
	if kID, ok := jwt.KeyID(); ok {
		keysFunc = c.keysFuncWithID(kID)
	} else {
		keysFunc = c.keysFuncAll()
	}

	v := oidc.NewJWTVerifier(
		c.providerConfig.Get().Issuer.String(),
		c.credentials.ID,
		c.maybeSyncKeys, keysFunc)

	return v.Verify(jwt)
}

func (c *Client) VerifyJWTForClientID(jwt jose.JWT, clientID string) error {
	var keysFunc func() []key.PublicKey
	if kID, ok := jwt.KeyID(); ok {
		keysFunc = c.keysFuncWithID(kID)
	} else {
		keysFunc = c.keysFuncAll()
	}

	v := oidc.NewJWTVerifier(
		c.providerConfig.Get().Issuer.String(),
		clientID,
		c.maybeSyncKeys, keysFunc)

	return v.Verify(jwt)
}

// keysFuncWithID returns a function that retrieves at most unexpired
// public key from the Client that matches the provided ID
func (c *Client) keysFuncWithID(kID string) func() []key.PublicKey {
	return func() []key.PublicKey {
		c.keySetSyncMutex.RLock()
		defer c.keySetSyncMutex.RUnlock()

		if c.keySet.ExpiresAt().Before(time.Now()) {
			return []key.PublicKey{}
		}

		k := c.keySet.Key(kID)
		if k == nil {
			return []key.PublicKey{}
		}

		return []key.PublicKey{*k}
	}
}

// keysFuncAll returns a function that retrieves all unexpired public
// keys from the Client
func (c *Client) keysFuncAll() func() []key.PublicKey {
	return func() []key.PublicKey {
		c.keySetSyncMutex.RLock()
		defer c.keySetSyncMutex.RUnlock()

		if c.keySet.ExpiresAt().Before(time.Now()) {
			return []key.PublicKey{}
		}

		return c.keySet.Keys()
	}
}

type providerConfigRepo struct {
	mu     sync.RWMutex
	config oidc.ProviderConfig // do not access directly, use Get()
}

func newProviderConfigRepo(pc oidc.ProviderConfig) *providerConfigRepo {
	return &providerConfigRepo{sync.RWMutex{}, pc}
}

// returns an error to implement ProviderConfigSetter
func (r *providerConfigRepo) Set(cfg oidc.ProviderConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = cfg
	return nil
}

func (r *providerConfigRepo) Get() oidc.ProviderConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.config
}

type ClientCredsTokenRefresher struct {
	Issuer     string
	OIDCClient *Client
}

func (c *ClientCredsTokenRefresher) Verify(jwt jose.JWT) (err error) {
	_, err = oidc.VerifyClientClaims(jwt, c.Issuer)
	return
}

func (c *ClientCredsTokenRefresher) Refresh() (jwt jose.JWT, err error) {
	if err = c.OIDCClient.Healthy(); err != nil {
		err = fmt.Errorf("unable to authenticate, unhealthy OIDC client: %v", err)
		return
	}

	jwt, err = c.OIDCClient.ClientCredsToken([]string{"openid"})
	if err != nil {
		err = fmt.Errorf("unable to verify auth code with issuer: %v", err)
		return
	}

	return
}

func NewClient(clientID, clientSecret, discovery, redirectURL string, tlsConfig *tls.Config) (*Client, *ClientCredsTokenManager) {
	cc := oidc.ClientCredentials{
		ID:     clientID,
		Secret: clientSecret,
	}
	httpClient := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}

	var cfg oidc.ProviderConfig
	var err error

	for {
		cfg, err = oidc.FetchProviderConfig(httpClient, discovery)
		if err == nil {
			break
		}

		sleep := 3 * time.Second
		log.Errorf("Failed fetching provider config, trying again in %v: %v", sleep, err)
		time.Sleep(sleep)
	}

	log.Infof("Fetched provider config from %s", discovery)

	ccfg := oidc.ClientConfig{
		HTTPClient:     httpClient,
		ProviderConfig: cfg,
		Credentials:    cc,
		RedirectURL:    redirectURL,
	}

	client, err := NewOIDCClient(ccfg)
	if err != nil {
		log.Fatalf("Unable to create Client: %v", err)
	}

	client.SyncProviderConfig(discovery)

	tm := NewClientCredsTokenManager(client, discovery)
	tm.Run()
	tm.WaitUntilInitialSync()

	return client, tm
}

func NewClientCredsTokenManager(client *Client, issuer string) *ClientCredsTokenManager {
	return &ClientCredsTokenManager{
		client: client,
		issuer: issuer,
		clock:  clockwork.NewRealClock(),
	}
}

type ClientCredsTokenManager struct {
	client *Client
	issuer string
	Token  jose.JWT

	clock           clockwork.Clock
	initialSyncDone bool
	initialSyncWait sync.WaitGroup
}

func (tm *ClientCredsTokenManager) WaitUntilInitialSync() {
	tm.initialSyncWait.Wait()
}

func (tm *ClientCredsTokenManager) Run() chan struct{} {
	stop := make(chan struct{})

	var next pcsStepper
	next = &pcsStepNext{aft: time.Duration(0)}

	tm.initialSyncWait.Add(1)
	go func() {
		for {
			select {
			case <-tm.clock.After(next.after()):
				next = next.step(tm.fetchToken)
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func (tm *ClientCredsTokenManager) fetchToken() (time.Duration, error) {
	if err := tm.client.Healthy(); err != nil {
		return 0, fmt.Errorf("unable to authenticate, unhealthy OIDC client: %v", err)
	}

	jwt, err := tm.client.ClientCredsToken(oidc.DefaultScope)
	if err != nil {
		return 0, err
	}
	claims, err := jwt.Claims()
	if err != nil {
		return 0, err
	}
	ident, err := oidc.IdentityFromClaims(claims)
	if err != nil {
		return 0, err
	}

	tm.Token = jwt

	if !tm.initialSyncDone {
		tm.initialSyncWait.Done()
		tm.initialSyncDone = true
	}

	return nextSyncAfter(ident.ExpiresAt, tm.clock), nil
}

type pcsStepFunc func() (time.Duration, error)

type pcsStepper interface {
	after() time.Duration
	step(pcsStepFunc) pcsStepper
}

type pcsStepNext struct {
	aft time.Duration
}

func (n *pcsStepNext) after() time.Duration {
	return n.aft
}

func (n *pcsStepNext) step(fn pcsStepFunc) (next pcsStepper) {
	ttl, err := fn()
	if err == nil {
		next = &pcsStepNext{aft: ttl}
		log.Debugf("Refreshed jwt, next attempt in %v", next.after())
	} else {
		next = &pcsStepRetry{aft: time.Second}
		log.Errorf("JWT refresh failed, retrying in %v: %v", next.after(), err)
	}
	return
}

type pcsStepRetry struct {
	aft time.Duration
}

func (r *pcsStepRetry) after() time.Duration {
	return r.aft
}

func (r *pcsStepRetry) step(fn pcsStepFunc) (next pcsStepper) {
	ttl, err := fn()
	if err == nil {
		next = &pcsStepNext{aft: ttl}
		log.Infof("JWT refresh no longer failing")
	} else {
		next = &pcsStepRetry{aft: timeutil.ExpBackoff(r.aft, time.Minute)}
		log.Errorf("JWT refresh still failing, retrying in %v: %v", next.after(), err)
	}
	return
}

func nextSyncAfter(exp time.Time, clock clockwork.Clock) time.Duration {
	if exp.IsZero() {
		return MaximumTokenRefreshInterval
	}

	t := exp.Sub(clock.Now()) / 2
	if t > MaximumTokenRefreshInterval {
		t = MaximumTokenRefreshInterval
	} else if t < MinimumTokenRefreshInterval {
		t = MinimumTokenRefreshInterval
	}

	return t
}
