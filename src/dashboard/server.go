package dashboard

import (
	"crypto/tls"
	"crypto/x509"
	"dashboard/storage"
	"io/ioutil"
	"net"

	"encoding/json"
	"errors"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	pb "github.com/otsimo/otsimopb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	roots *x509.CertPool
)

type Server struct {
	Config       *CommandConfig
	Storage      storage.Driver
	Client       *Client
	TokenManager *ClientCredsTokenManager
	sc           *ServiceConfig
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

	pb.RegisterDashboardServiceServer(grpcServer, s)
	log.Infof("server.go: Binding %s for grpc", grpcPort)
	//Serve
	grpcServer.Serve(lis)
}

func NewServer(config *CommandConfig, driver storage.Driver) *Server {
	sc, err := readConfig(config.ConfigPath)
	if err != nil {
		panic(err)
	}
	server := &Server{
		Config:  config,
		Storage: driver,
		sc:      sc,
	}
	var tlsConfig tls.Config
	if config.TrustedCAFile != "" {
		roots = x509.NewCertPool()
		pemBlock, err := ioutil.ReadFile(config.TrustedCAFile)
		if err != nil {
			log.Fatalf("Unable to read ca file: %v", err)
		}
		roots.AppendCertsFromPEM(pemBlock)
		tlsConfig.RootCAs = roots
	}
	if config.NoAuth {
		log.Debugln("Creating new oidc client, discovery=", config.AuthDiscovery)
		client, tokenMan := NewClient(config.ClientID, config.ClientSecret, config.AuthDiscovery, "", &tlsConfig)
		server.Client = client
		server.TokenManager = tokenMan
	}
	if config.WatchConfigFile {
		go watchFile(config.ConfigPath)
	}
	return server
}

func readConfig(configPath string) (*ServiceConfig, error) {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Errorf("failed to read configuration file, %#v", err)
		return nil, err
	}
	desc := &ServiceConfig{}
	if filepath.Ext(configPath) == ".yaml" || filepath.Ext(configPath) == ".yml" {
		err = yaml.Unmarshal(data, desc)
	} else if filepath.Ext(configPath) == ".json" {
		err = json.Unmarshal(data, desc)
	} else {
		err = errors.New("unknwon data format")
	}
	if err != nil {
		log.Errorf("failed to unmarshal configuration file, %#v", err)
		return nil, err
	}
	return desc, nil
}
