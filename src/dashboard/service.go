package dashboard

import (
	"time"

	"dashboard/storage"

	"github.com/Sirupsen/logrus"
	pb "github.com/otsimo/otsimopb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type taskResult struct {
	success  bool
	provider string
	items    *pb.ProviderItems
	cache    []*storage.Item
}

func (d *Server) workerSyncDB(p *Provider, req pb.ProviderGetRequest, timeout int64, results chan<- taskResult) {
	c1 := make(chan taskResult, 1)
	go func() {
		items, err := d.Storage.GetItems(req.Request.ProfileId, p.config.Name, req.Request.LastTimeDataFetched)
		if err != nil {
			logrus.Errorf("failed to get items from %s db cache, err=%v", p.config.Name, err)
			c1 <- taskResult{success: false, provider: p.config.Name}
			return
		}
		logrus.Debugf("service.go:dbworker: get results from cache provider %s, count=%d", p.config.Name, len(items))
		c1 <- taskResult{success: true, cache: items, provider: p.config.Name}
	}()
	select {
	case res := <-c1:
		results <- res
	case <-time.After(time.Millisecond * time.Duration(timeout)):
		logrus.Errorf("service.go:dbworker: timeout, failed to get result from provider %s", p.config.Name)
		c1 <- taskResult{success: false, provider: p.config.Name}
	}
}

func workerSync(p *Provider, req pb.ProviderGetRequest, timeout int64, results chan<- taskResult) {
	client := p.Get()
	if client == nil {
		p.Close()
		client = p.Get()
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*time.Duration(timeout))
	pi, err := client.Get(ctx, &req)
	if err != nil {
		logrus.Errorf("failed to get items from %s, err=%v", p.config.Name, err)
		results <- taskResult{success: false, provider: p.config.Name}
	} else {
		results <- taskResult{success: true, items: pi, provider: p.config.Name}
	}
}

func (d *Server) processResultSync(to *pb.DashboardItems, req *pb.DashboardGetRequest, res taskResult) {
	if !res.success {
		return
	}
	//todo(sercan) cache result
	name := ""
	score := float32(1.0)
	for _, v := range d.providers {
		if v.name == res.provider {
			v.configLock.RLock()
			name = v.name
			score = v.config.ScoreMultiplier
			v.configLock.RUnlock()
		}
	}
	for _, v := range res.items.Items {
		item := v.Item
		item.ProviderWeight = score
		item.ProviderName = name
		//todo(sercan) possible race?
		to.Items = append(to.Items, item)
	}
	for _, v := range res.cache {
		item := v.GetCard()
		if item == nil {
			continue
		}
		item.ProviderWeight = score
		item.ProviderName = name
		to.Items = append(to.Items, item)
	}
}

func (d *Server) Get(ctx context.Context, in *pb.DashboardGetRequest) (*pb.DashboardItems, error) {
	logrus.Infof("service.go:GET: %+v", in)
	uinfo, err := checkContext(ctx, d.Client)
	if err != nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "PermissionDenied:%v", err)
	}
	if in.ProfileId != uinfo.UserID {
		return nil, grpc.Errorf(codes.PermissionDenied, "PermissionDenied: User cannot see others dashboard")
	}
	usr := d.Storage.GetUser(in.ProfileId)
	//todo(sercan) filter providers by users info,
	n := len(d.providers)

	results := make(chan taskResult, n)
	defer close(results)
	req := pb.ProviderGetRequest{
		Request:    in,
		UserGroups: []string{uinfo.UserGroup},
	}
	now := time.Now().Unix()
	for _, v := range d.providers {
		fromDB := false
		for _, p := range usr.Providers {
			if p.Name == v.config.Name {
				if p.ExpiresAt.Unix() > now {
					fromDB = true
				}
			}
		}
		if fromDB {
			go d.workerSyncDB(v, req, 1000, results)
		} else {
			go workerSync(v, req, 2000, results)
		}
	}
	res := &pb.DashboardItems{
		ProfileId: in.ProfileId,
		ChildId:   in.ChildId,
		CreatedAt: time.Now().Unix(),
		Items:     []*pb.Card{},
	}
	for a := 1; a <= n; a++ {
		r := <-results
		d.processResultSync(res, in, r)
	}
	logrus.Debugf("service.go: send result to client: count=%d", len(res.Items))
	return res, nil
}
