package dashboard

import (
	"crypto/tls"
	"crypto/x509"
	"dashboard/storage"
	"io/ioutil"
	"net"

	log "github.com/Sirupsen/logrus"
	pb "github.com/otsimo/otsimopb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	Config       *Config
	Storage      storage.Driver
	Client       *Client
	TokenManager *ClientCredsTokenManager
}

func (s *Server) ListenGRPC() {
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

	dashboard := &dashboardGrpcService{
		server: s,
	}

	pb.RegisterDashboardServiceServer(grpcServer, dashboard)
	log.Infof("server.go: Binding %s for grpc", grpcPort)
	//Serve
	grpcServer.Serve(lis)
}

func NewServer(config *Config, driver storage.Driver) *Server {
	server := &Server{
		Config:  config,
		Storage: driver,
	}
	log.Debugln("Creating new oidc client discovery=", config.AuthDiscovery)
	var tlsConfig tls.Config
	var roots *x509.CertPool
	if config.TrustedCAFile != "" {
		roots = x509.NewCertPool()
		pemBlock, err := ioutil.ReadFile(config.TrustedCAFile)
		if err != nil {
			log.Fatalf("Unable to read ca file: %v", err)
		}
		roots.AppendCertsFromPEM(pemBlock)
		tlsConfig.RootCAs = roots
	}
	client, tokenMan := NewClient(config.ClientID, config.ClientSecret, config.AuthDiscovery, "", &tlsConfig)
	server.Client = client
	server.TokenManager = tokenMan

	jwtCreds := NewOauthAccess(tokenMan)

	var opts []grpc.DialOption
	if roots != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(roots, "")))
	} else {
		jwtCreds.RequireTLS = false
		opts = append(opts, grpc.WithInsecure())
	}
	opts = append(opts, grpc.WithPerRPCCredentials(&jwtCreds))
	return server
}
