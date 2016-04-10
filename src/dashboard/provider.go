package dashboard

import (
	"github.com/Sirupsen/logrus"
	"github.com/otsimo/otsimopb"
	"google.golang.org/grpc"
)

type ProviderConfig struct {
	Name               string                 `json:"name"`
	ServiceURL         string                 `json:"url"`
	ScoreMultiplier    float32                `json:"score"`
	InsecureConnection bool                   `json:"insecure"`
	RequiresAuth       bool                   `json:"requiresAuth"`
	Info               *otsimopb.ProviderInfo `json:"-"`
}

type Provider struct {
	config     ProviderConfig
	connection *grpc.Conn
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
	aconn, err := grpc.Dial(ac.config.ServiceURL, nil)
	if err != nil {
		logrus.Fatalf("provider.go: did not connect to remote provider service: %v", err)
	}
	ac.client = otsimopb.NewDashboardProviderClient(aconn)
	ac.connection = aconn
	return ac.client
}
