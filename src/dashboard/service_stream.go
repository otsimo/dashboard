package dashboard

import (
	"time"

	"github.com/Sirupsen/logrus"
	pb "github.com/otsimo/otsimopb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type DashboardStream interface {
	Send(*pb.Card) error
}

func workerAsync(p *Provider, req pb.ProviderGetRequest, timeout int64, results chan<- bool, stream DashboardStream) {
	//todo(sercan) look for caches, if a valid cached request is valid return it
	client := p.Get()
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	pi, err := client.Get(ctx, &req)
	if err != nil {
		logrus.Errorf("service_stream.go:worker: failed to get items from %s, err=%v", p.config.Name, err)
		results <- false
		return
	}
	logrus.Debugf("service_stream.go:worker: get results from provider %s, count=%d", p.config.Name, len(pi.Items))
	for _, v := range pi.Items {
		p.configLock.RLock()
		v.Item.ProviderWeight = p.config.ScoreMultiplier
		p.configLock.RUnlock()

		v.Item.ProviderName = p.Name()
		stream.Send(v.Item)
	}
	results <- true
}

func (d *Server) GetStream(in *pb.DashboardGetRequest, stream pb.DashboardService_GetStreamServer) error {
	logrus.Infof("service.go:GET_STREAM: %+v", in)
	uinfo, err := checkContext(stream.Context(), d.Client)
	if err != nil {
		return grpc.Errorf(codes.PermissionDenied, "PermissionDenied: %v", err)
	}
	if in.ProfileId != uinfo.UserID {
		return grpc.Errorf(codes.PermissionDenied, "PermissionDenied: User cannot see others dashboard")
	}
	n := len(d.providers)
	results := make(chan bool, n)
	defer close(results)

	req := pb.ProviderGetRequest{
		Request:    in,
		UserGroups: []string{uinfo.UserGroup},
	}
	for _, v := range d.providers {
		go workerAsync(v, req, 2000, results, stream)
	}

	for a := 1; a <= n; a++ {
		<-results
	}
	return nil
}
