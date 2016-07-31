package dashboard

import (
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/otsimo/otsimopb"
	"github.com/sercand/kuberesolver"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ProviderConfig struct {
	Name               string  `json:"name"`
	ServiceURL         string  `json:"url"`
	ScoreMultiplier    float32 `json:"score"`
	InsecureConnection bool    `json:"insecure"`
	RequiresAuth       bool    `json:"auth"`
	info               *otsimopb.ProviderInfo
}

type Provider struct {
	config     ProviderConfig
	connection *grpc.ClientConn
	client     otsimopb.DashboardProviderClient
	configLock sync.RWMutex
	name       string
	balancer   *kuberesolver.Balancer
}

func (ac *Provider) Close() {
	if ac.connection != nil {
		ac.connection.Close()
	}
	ac.connection = nil
}

func (ac *Provider) Get() otsimopb.DashboardProviderClient {
	if ac.connection != nil {
		return ac.client
	}
	ac.configLock.RLock()

	logrus.Infof("provider.go:[%s]: connecting to %s service provider", ac.config.Name, ac.config.Name)
	var opts []grpc.DialOption
	if ac.config.InsecureConnection {
		opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(roots, "")))
	}
	conn, err := ac.balancer.Dial(ac.config.ServiceURL, opts...)
	if err != nil {
		logrus.Errorf("provider.go: did not connect to remote provider service: %v", err)
		return nil
	}
	logrus.Infof("provider.go:[%s]: connected to %s service provider", ac.config.Name, ac.config.Name)

	ac.configLock.RUnlock()

	ac.client = otsimopb.NewDashboardProviderClient(conn)
	ac.connection = conn
	return ac.client
}

func (ac *Provider) Init() {
	clt := ac.Get()
	//	if ac.config.RequiresAuth {
	if clt == nil {
		logrus.Errorf("provider.go:[%s]: client is nil", ac.name)
		return
	}
	logrus.Infof("provider.go:[%s]: init calling", ac.name)
	pi, err := clt.Info(context.Background(), &otsimopb.ProviderInfoRequest{})
	if err == nil {
		logrus.Infof("provider.go:[%s]: info=%+v", ac.name, pi)
		ac.configLock.Lock()
		ac.config.info = pi
		ac.configLock.Unlock()
	} else {
		logrus.Errorf("provider.go:[%s]: failed to get info, err=%v", ac.name, err)
	}
}

func (ac *Provider) ReInit() {
	ac.Close()
	clt := ac.Get()
	//	if ac.config.RequiresAuth {
	logrus.Infof("provider.go:[%s]: reinit calling", ac.name)
	pi, err := clt.Info(context.Background(), &otsimopb.ProviderInfoRequest{})
	if err == nil {
		logrus.Infof("provider.go:[%s]: info=%+v", ac.name, pi)
		ac.configLock.Lock()
		ac.config.info = pi
		ac.configLock.Unlock()
	} else {
		logrus.Errorf("provider.go:[%s]: failed to get info, err=%v", ac.name, err)
	}
}

func (ac *Provider) MergeConfig(config ProviderConfig) {
	ac.configLock.Lock()
	ac.config = config
	ac.configLock.Unlock()
	go ac.ReInit()
}

func (ac *Provider) Name() string {
	return ac.name
}

func NewProvider(cnf ProviderConfig, balancer *kuberesolver.Balancer) *Provider {
	return &Provider{
		config:   cnf,
		name:     cnf.Name,
		balancer: balancer,
	}
}
