package dashboard

import (
	"errors"

	pb "github.com/otsimo/otsimopb"
	"golang.org/x/net/context"
)

func (d *Server) Get(ctx context.Context, in *pb.DashboardGetRequest) (*pb.DashboardItems, error) {

	return nil, errors.New("not implemented")
}
