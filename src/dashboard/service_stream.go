package dashboard

import (
	"time"

	"github.com/Sirupsen/logrus"
	pb "github.com/otsimo/otsimopb"
	"golang.org/x/net/context"
)

func workerAsync(p *Provider, req pb.DashboardGetRequest, timeout int64, results chan<- taskResult, done <-chan bool, stream pb.DashboardService_GetStreamServer) {
	//todo(sercan) look for caches, if a valid cached request is valid return it
	client := p.Get()
	if client == nil {
		p.Close()
		client = p.Get()
	}
	c1 := make(chan taskResult, 1)
	go func() {
		pi, err := client.Get(context.Background(), &req)
		if err != nil {
			logrus.Errorf("failed to get items from %s, err=%v", p.config.Name, err)
			c1 <- taskResult{success: false, provider: p.config.Name}
			return
		}
		logrus.Debugf("service.go:worker: get results from provider %s, count=%d", p.config.Name, len(pi.Items))
		for _, v := range pi.Items {
			v.Item.ProviderWeight = p.config.ScoreMultiplier
			v.Item.ProviderName = p.config.Name
			stream.Send(v.Item)
		}
		c1 <- taskResult{success: true, provider: p.config.Name}
	}()
	select {
	case res := <-c1:
		results <- res
	case <-time.After(time.Millisecond * time.Duration(timeout)):
		logrus.Errorf("service.go:worker: timeout, failed to get result from provider %s", p.config.Name)
		c1 <- taskResult{success: false, provider: p.config.Name}
	case <-done:
		c1 <- taskResult{success: false, provider: p.config.Name}
	}
}

func (d *Server) GetStream(in *pb.DashboardGetRequest, stream pb.DashboardService_GetStreamServer) error {
	logrus.Infof("service.go:GET_STREAM: %+v", in)

	//todo(sercan) filter providers by users info,
	n := len(d.providers)

	results := make(chan taskResult, n)
	defer close(results)
	done := make(chan bool, 1)
	defer close(done)

	for _, v := range d.providers {
		go workerAsync(v, *in, 10000, results, done, stream)
	}

	for a := 1; a <= n; a++ {
		<-results
	}
	return nil
}
