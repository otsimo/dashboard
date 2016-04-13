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
	logrus.Infof("provider.go:[%s]: connecting to %s service provider", ac.config.Name, ac.config.Name)

	var opts []grpc.DialOption
	if ac.config.InsecureConnection {
		opts = append(opts, grpc.WithInsecure())
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(roots, "")))
	}

	conn, err := grpc.Dial(ac.config.ServiceURL, opts...)
	if err != nil {
		logrus.Errorf("provider.go: did not connect to remote provider service: %v", err)
		return nil
	}
	logrus.Infof("provider.go:[%s]: connected to %s service provider", ac.config.Name, ac.config.Name)
	ac.client = otsimopb.NewDashboardProviderClient(conn)
	ac.connection = conn
	return ac.client
}

func (ac *Provider) Init() {
	clt := ac.Get()
	//	if ac.config.RequiresAuth {
	logrus.Infof("provider.go:[%s]: init calling", ac.config.Name)
	pi, err := clt.Info(context.Background(), &otsimopb.ProviderInfoRequest{})
	if err == nil {
		logrus.Infof("provider.go:[%s]: info=%+v", ac.config.Name, pi)
		ac.config.info = pi
	} else {
		logrus.Errorf("provider.go:[%s]: failed to get info, err=%v", ac.config.Name, err)
	}
}
