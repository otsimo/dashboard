package dashboard

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
)

type OIDCCredentials interface {
	Credentials() credentials.PerRPCCredentials
}

// oauthAccess supplies credentials from a given token.
type oauthAccess struct {
	tm         *ClientCredsTokenManager
	RequireTLS bool
}

// NewOauthAccess constructs the credentials using a given token.
func NewOauthAccess(tm *ClientCredsTokenManager) credentials.PerRPCCredentials {
	return &oauthAccess{tm: tm, RequireTLS: true}
}

func (oa *oauthAccess) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": fmt.Sprintf("Bearer %s", oa.tm.Token.Encode()),
	}, nil
}

func (oa *oauthAccess) RequireTransportSecurity() bool {
	return oa.RequireTLS
}
