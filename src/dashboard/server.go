package dashboard

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	"storage"

	log "github.com/Sirupsen/logrus"
	pb "github.com/otsimo/api/apipb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

type Server struct {
	Config       *Config
	Storage      storage.Driver
	Client       *Client
	TokenManager *ClientCredsTokenManager
	Api          pb.ApiServiceClient
	Analysis     pb.AnalysisServiceClient
	Registry     pb.RegistryServiceClient

	apiConn      *grpc.ClientConn
	analysisConn *grpc.ClientConn
	registryConn *grpc.ClientConn
}

func (s *Server) ListenGRPC() {
	grpcPort := s.Config.GetGrpcPortString()
	//Listen
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("server.go: failed to listen %v for grpc", err)
	}
	var l = &log.Logger{
		Out:       os.Stdout,
		Formatter: &log.TextFormatter{FullTimestamp: true},
		Hooks:     make(log.LevelHooks),
		Level:     log.GetLevel(),
	}
	grpclog.SetLogger(l)

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

	// API
	conn, err := grpc.Dial(config.ApiServiceURL, opts...)
	if err != nil {
		log.Fatalf("server.go: did not connect to api service: %v", err)
	}
	server.Api = pb.NewApiServiceClient(conn)
	server.apiConn = conn

	// ANALYSIS
	aconn, err := grpc.Dial(config.AnalysisServiceURL, opts...)
	if err != nil {
		log.Fatalf("server.go: did not connect to analysis service: %v", err)
	}
	server.Analysis = pb.NewAnalysisServiceClient(aconn)
	server.analysisConn = aconn

	// REGISTRY
	rconn, err := grpc.Dial(config.RegistryServiceURL, opts...)
	if err != nil {
		log.Fatalf("server.go: did not connect to registry service: %v", err)
	}
	server.Registry = pb.NewRegistryServiceClient(rconn)
	server.registryConn = rconn

	return server
}

// oauthAccess supplies credentials from a given token.
type oauthAccess struct {
	tm         *ClientCredsTokenManager
	RequireTLS bool
}

// NewOauthAccess constructs the credentials using a given token.
func NewOauthAccess(tm *ClientCredsTokenManager) oauthAccess {
	return oauthAccess{tm: tm, RequireTLS: true}
}

func (oa *oauthAccess) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": "Bearer" + " " + oa.tm.Token.Encode(),
	}, nil
}

func (oa *oauthAccess) RequireTransportSecurity() bool {
	return oa.RequireTLS
}
