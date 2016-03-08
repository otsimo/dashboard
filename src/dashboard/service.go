package dashboard

import (
	"errors"

	"github.com/otsimo/api/apipb"
	"golang.org/x/net/context"
)

type dashboardGrpcService struct {
	server *Server
}

func (d *dashboardGrpcService) Get(ctx context.Context, in *apipb.DashboardGetRequest) (*apipb.Dashboard, error) {

	return nil, errors.New("not implemented")
}
