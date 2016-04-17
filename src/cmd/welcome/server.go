package main

import (
	"crypto/tls"
	"crypto/x509"
	"dashboard"
	"errors"
	"io/ioutil"
	"net"

	"time"

	log "github.com/Sirupsen/logrus"
	pb "github.com/otsimo/otsimopb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	OneDay = 60 * 60 * 24
	OneWeek = 7 * OneDay
)

type Server struct {
	Config *Config
	client *dashboard.Client
	tm     *dashboard.ClientCredsTokenManager
	api    *LazyApiClient
}

func NewServer(config *Config) *Server {
	s := &Server{Config: config}
	s.Creds(true)
	return s
}

func (s *Server) Listen() {
	grpcPort := s.Config.GetGrpcPortString()
	//Listen
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("server.go: failed to listen %v for grpc", err)
	}
	var opts []grpc.ServerOption
	if s.Config.TlsCertFile != "" && s.Config.TlsKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(s.Config.TlsCertFile, s.Config.TlsKeyFile)
		if err != nil {
			log.Fatalf("server.go: Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)

	//register server
	pb.RegisterDashboardProviderServer(grpcServer, s)
	log.Infof("server.go: Binding %s for grpc", grpcPort)
	//Serve
	grpcServer.Serve(lis)
}

func (s *Server) Creds(getToken bool) {
	var roots *x509.CertPool
	tlsConfig := tls.Config{ServerName: ""}
	var opts []grpc.DialOption

	if s.Config.TrustedCAFile != "" {
		roots = x509.NewCertPool()
		pemBlock, err := ioutil.ReadFile(s.Config.TrustedCAFile)
		if err != nil {
			log.Fatalf("Unable to read ca file: %v", err)
		}
		roots.AppendCertsFromPEM(pemBlock)
		tlsConfig.RootCAs = roots
	}

	if s.Config.ApiConnectMode == "insecure-tls" {
		tlsConfig.InsecureSkipVerify = true
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig)))
	} else if s.Config.ApiConnectMode == "tls" {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tlsConfig)))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if getToken {
		log.Debugln("Creating new oidc client discovery=", s.Config.AuthDiscovery)
		s.client, s.tm = dashboard.NewClient(s.Config.ClientID, s.Config.ClientSecret, s.Config.AuthDiscovery, "", &tlsConfig)
		opts = append(opts, grpc.WithPerRPCCredentials(dashboard.NewOauthAccess(s.tm)))
	}

	s.api = NewLazyApiClient(s.Config.ApiServiceURL, opts)
}

func (d *Server) Info(ctx context.Context, in *pb.ProviderInfoRequest) (*pb.ProviderInfo, error) {
	log.Infof("server.go:INFO: %+v", in)
	api := d.api.Get()
	log.Infof("server.go:INFO: connected to api service %v", api)
	return &pb.ProviderInfo{}, nil
}

func NewItem(in *pb.DashboardGetRequest, p *pb.Profile, delta, score, id int64) *pb.ProviderItem {
	ttl := OneWeek - delta
	card := NewCard(in, int64(ttl), p, id)
	card.ProviderScore = int32(score)
	if score > 495 {
		card.Decoration.Size_ = pb.LARGE
	} else if score > 450 {
		card.Decoration.Size_ = pb.MEDIUM
	} else {
		card.Decoration.Size_ = pb.SMALL
	}
	return &pb.ProviderItem{
		Cacheable: true,
		Ttl:       ttl,
		Item:      card,
	}
}

func (d *Server) Get(ctx context.Context, in *pb.DashboardGetRequest) (*pb.ProviderItems, error) {
	log.Infof("server.go:GET: %+v", in)
	api := d.api.Get()
	p, err := api.GetProfile(context.Background(), &pb.GetProfileRequest{Id: in.ProfileId})
	if err != nil {
		log.Errorf("server.go:GET: profile not found")
		return nil, errors.New("Not found")
	}
	now := time.Now().Unix()

	res := &pb.ProviderItems{
		ProfileId: in.ProfileId,
		ChildId:   in.ChildId,
		CreatedAt: now,
	}
	delta := now - p.CreatedAt
	if delta < 0 {
		delta = now - (p.CreatedAt / 1e3)
	}

	if delta < OneWeek {
		res.Cacheable = true
		res.Ttl = OneWeek - delta
		res.Items = make([]*pb.ProviderItem, 1)
		res.Items[0] = NewItem(in, p, delta, 490, 1)
	} else {
		res.Cacheable = true
		res.Ttl = OneWeek * 4
		res.Items = []*pb.ProviderItem{}
	}
	return res, nil
}
