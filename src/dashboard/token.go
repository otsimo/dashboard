package dashboard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/go-oidc/jose"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

const (
	OtsimoUserTypeClaim = "otsimo.com/typ"
	OtsimoAdminUserType = "adm"
)

type userInfo struct {
	UserID    string
	Token     jose.JWT
	UserGroup string
}

func getJWTToken(ctx context.Context) (jose.JWT, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return jose.JWT{}, fmt.Errorf("missing metadata")
	}
	var auth []string
	auth, ok = md["authorization"]
	if !ok || len(auth) == 0 {
		auth, ok = md["Authorization"]
		if !ok || len(auth) == 0 {
			return jose.JWT{}, errors.New("missing authorization header")
		}
	}

	ah := auth[0]
	if len(ah) <= 6 || strings.ToUpper(ah[0:6]) != "BEARER" {
		return jose.JWT{}, errors.New("should be a bearer token")
	}
	val := ah[7:]
	if len(val) == 0 {
		return jose.JWT{}, errors.New("bearer token is empty")
	}
	return jose.ParseJWT(val)
}

func checkContext(ctx context.Context, client *Client) (*userInfo, error) {
	jwt, err := getJWTToken(ctx)
	if err != nil {
		return nil, err
	}
	claims, err := jwt.Claims()
	if err != nil {
		return nil, fmt.Errorf("token.go: failed to get claims %v", err)
	}

	sub, ok, err := claims.StringClaim("sub")
	if err != nil {
		return nil, fmt.Errorf("token.go: failed to parse 'sub' claim: %v", err)
	}
	if !ok || sub == "" {
		return nil, errors.New("token.go: missing required 'sub' claim")
	}

	aud, _, _ := claims.StringClaim("aud")
	typ, _, _ := claims.StringClaim(OtsimoUserTypeClaim)
	ug := "otsimo.com/user"
	if typ == OtsimoAdminUserType {
		ug = "otsimo.com/admin"
	}
	err = client.VerifyJWTForClientID(jwt, aud)
	if err != nil {
		logrus.Errorf("failed to verify jwt, error=%v", err)
		return nil, err
	}
	return &userInfo{
		UserID:    sub,
		UserGroup: ug,
		Token:     jwt,
	}, nil
}
