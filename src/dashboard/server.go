package dashboard

import (
	"crypto/tls"
	"crypto/x509"
	"dashboard/storage"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"path/filepath"
	"time"

	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/ghodss/yaml"
	"github.com/otsimo/health"
	tlsChecker "github.com/otsimo/health/tls"
	pb "github.com/otsimo/otsimopb"
	"github.com/sercand/kuberesolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
)

var (
	roots *x509.CertPool
)

type Server struct {
	Config       *CommandConfig
	Storage      storage.Driver
	Client       *Client
	TokenManager *ClientCredsTokenManager
	providers    []*Provider
	balancer     *kuberesolver.Balancer
	tlsChecker   *tlsChecker.TLSHealthChecker
}

func (s *Server) Healthy() error {
	if s.tlsChecker != nil {
		return s.tlsChecker.Healthy()
	}
	return nil
}

func (s *Server) Listen() error {
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
		s.tlsChecker = tlsChecker.New(s.Config.TlsCertFile, s.Config.TlsKeyFile, time.Hour*24*21)
	}

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterDashboardServiceServer(grpcServer, s)
	hs := health.New(s.Storage, s)
	grpc_health_v1.RegisterHealthServer(grpcServer, hs)

	log.Infof("server.go: Binding %s for grpc", grpcPort)
	go http.ListenAndServe(s.Config.GetHealthPortString(), hs)
	//Serve
	return grpcServer.Serve(lis)
}

func NewServer(config *CommandConfig, driver storage.Driver) *Server {
	sc, err := readConfig(config.ConfigPath)
	if err != nil {
		panic(err)
	}
	server := &Server{
		Config:   config,
		Storage:  driver,
		balancer: kuberesolver.New(),
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

	if !config.NoAuth {
		log.Debugln("Creating new oidc client, discovery=", config.AuthDiscovery)
		client, tokenMan := NewClient(config.ClientID, config.ClientSecret, config.AuthDiscovery, "", &tlsConfig)
		server.Client = client
		server.TokenManager = tokenMan
	}
	server.providers = make([]*Provider, len(sc.Providers))
	for i, v := range sc.Providers {
		server.providers[i] = NewProvider(v, server.balancer)
	}
	if config.WatchConfigFile {
		go watchFile(config.ConfigPath, server)
	}
	go server.InitProviders()
	return server
}

func readConfig(configPath string) (*ServiceConfig, error) {
	maxNumberOfRetry := 3
	var err error
	var data []byte
	for i := 0; i < maxNumberOfRetry; i++ {
		data, err = ioutil.ReadFile(configPath)
		if err != nil {
			log.Errorf("failed to read configuration file, %+v", err)
			time.Sleep(time.Second * time.Duration(5*(i+1)))
			continue
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
			log.Errorf("failed to unmarshal configuration file, %+v", err)
			time.Sleep(time.Second * time.Duration(5*(i+1)))
			continue
		}
		return desc, nil
	}
	return nil, err
}

func (s *Server) InitProviders() {
	for _, v := range s.providers {
		go v.Init()
	}
}

func (s *Server) RereadConfig() {
	sc, err := readConfig(s.Config.ConfigPath)
	if err != nil {
		log.Errorf("failed to read config file err=%v", err)
		return
	}
	for _, v := range sc.Providers {
		founded := false
		for _, sp := range s.providers {
			if sp.Name() == v.Name {
				sp.MergeConfig(v)
				founded = true
			}
		}
		if !founded {
			p := NewProvider(v, s.balancer)
			s.providers = append(s.providers, p)
			go p.Init()
		}
	}
}
