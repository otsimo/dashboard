package dashboard

import (
	"github.com/Sirupsen/logrus"
	"github.com/otsimo/otsimopb"
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

	var opts []grpc.DialOption
	if ac.config.InsecureConnection {
		opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(roots, "")))
	}

	conn, err := grpc.Dial(ac.config.ServiceURL, opts...)
	if err != nil {
		logrus.Fatalf("provider.go: did not connect to remote provider service: %v", err)
	}
	ac.client = otsimopb.NewDashboardProviderClient(conn)
	ac.connection = conn
	return ac.client
}

func (ac *Provider) Init() {
	clt := ac.Get()
	//	if ac.config.RequiresAuth {
	pi, err := clt.Info(context.Background(), otsimopb.ProviderInfoRequest{})
	if err != nil {
		ac.config.info = pi
	}
}
