package dashboard

import (
	"time"

	"github.com/Sirupsen/logrus"
	pb "github.com/otsimo/otsimopb"
	"golang.org/x/net/context"
)

type taskResult struct {
	success  bool
	provider string
	items    *pb.ProviderItems
}

func worker(p *Provider, req pb.DashboardGetRequest, timeout int64, results chan <- taskResult) {
	//todo(sercan) look for caches, if a valid cached request is valid return it
	client := p.Get()
	c1 := make(chan taskResult, 1)
	go func() {
		pi, err := client.Get(context.Background(), &req)
		if err != nil {
			logrus.Errorf("failed to get items from %s, err=%v", p.config.Name, err)
			c1 <- taskResult{success: false, provider: p.config.Name}
			return
		}
		c1 <- taskResult{success: true, items: pi, provider: p.config.Name}
	}()
	select {
	case res := <-c1:
		results <- res
	case <-time.After(time.Millisecond * time.Duration(timeout)):
		c1 <- taskResult{success: false, provider: p.config.Name}
	}
}

func (d *Server) processResult(to *pb.DashboardItems, req *pb.DashboardGetRequest, res taskResult) {
	if !res.success {
		return
	}
	//todo(sercan) cache result
	var pr *ProviderConfig
	for _, v := range d.providers {
		if v.config.Name == res.provider {
			pr = &v.config
		}
	}
	for _, v := range res.items.Items {
		item := v.Item
		item.ProviderWeight = pr.ScoreMultiplier
		//todo(sercan) possible race?
		to.Items = append(to.Items, item)
	}
}

func (d *Server) Get(ctx context.Context, in *pb.DashboardGetRequest) (*pb.DashboardItems, error) {
	//todo(sercan) filter providers by users info,
	n := len(d.providers)

	results := make(chan taskResult, n)
	defer close(results)

	for _, v := range d.providers {
		go worker(v, *in, 700, results)
	}
	res := &pb.DashboardItems{
		ProfileId: in.ProfileId,
		ChildId:   in.ChildId,
		CreatedAt: time.Now().Unix(),
		Items:     []*pb.Card{},
	}
	for a := 1; a <= n; a++ {
		r := <-results
		d.processResult(res, in, r)
	}
	return res, nil
}
