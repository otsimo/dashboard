package dashboard

import (
	"errors"

	pb "github.com/otsimo/otsimopb"
	"golang.org/x/net/context"
)

type dashboardGrpcService struct {
	server *Server
}

func (d *dashboardGrpcService) Get(ctx context.Context, in *pb.DashboardGetRequest) (*pb.DashboardItems, error) {

	return nil, errors.New("not implemented")
}
