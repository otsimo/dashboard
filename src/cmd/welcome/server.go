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
	OneDay  = 60 * 60 * 24
	OneWeek = 8 * OneDay
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
	log.Debugln("Creating new oidc client discovery=", s.Config.AuthDiscovery)
	var tlsConfig tls.Config
	var roots *x509.CertPool
	if s.Config.TrustedCAFile != "" {
		roots = x509.NewCertPool()
		pemBlock, err := ioutil.ReadFile(s.Config.TrustedCAFile)
		if err != nil {
			log.Fatalf("Unable to read ca file: %v", err)
		}
		roots.AppendCertsFromPEM(pemBlock)
		tlsConfig.RootCAs = roots
	}

	var opts []grpc.DialOption
	if roots != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(roots, "")))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if getToken {
		s.client, s.tm = dashboard.NewClient(s.Config.ClientID, s.Config.ClientSecret, s.Config.AuthDiscovery, "", &tlsConfig)
		jwtCreds := dashboard.NewOauthAccess(s.tm)
		opts = append(opts, grpc.WithPerRPCCredentials(jwtCreds))
	}

	s.api = NewLazyApiClient(s.Config.ApiServiceURL, opts)
}

func (d *Server) Info(ctx context.Context, in *pb.ProviderInfoRequest) (*pb.ProviderInfo, error) {
	return &pb.ProviderInfo{}, nil
}

func (d *Server) Get(ctx context.Context, in *pb.DashboardGetRequest) (*pb.ProviderItems, error) {
	api := d.api.Get()
	p, err := api.GetProfile(context.Background(), &pb.GetProfileRequest{Id: in.ProfileId})
	if err != nil {
		return nil, errors.New("Not found")
	}
	now := time.Now().Unix()

	res := &pb.ProviderItems{
		ProfileId: in.ProfileId,
		ChildId:   in.ChildId,
		CreatedAt: now,
	}
	if now-p.CreatedAt < OneWeek {
		res.Cacheable = true
		res.Ttl = now - p.CreatedAt
		res.Items = make([]*pb.ProviderItem, 1)
		pit := &pb.ProviderItem{
			Cacheable: true,
			Ttl:       now - p.CreatedAt,
			Item:      NewCard(in, res.Ttl),
		}
		res.Items[0] = pit
	} else {
		res.Cacheable = false
		res.Ttl = 0
		res.Items = []*pb.ProviderItem{}
	}
	return res, nil
}
