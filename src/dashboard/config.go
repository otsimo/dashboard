package dashboard

import "fmt"

const (
	DefaultGrpcPort = 18860
)

type Config struct {
	Debug         bool
	GrpcPort      int
	TlsCertFile   string
	TlsKeyFile    string
	ClientID      string
	ClientSecret  string
	TrustedCAFile string
	AuthDiscovery string
	ConfigPath    string
}

func (c *Config) GetGrpcPortString() string {
	return fmt.Sprintf(":%d", c.GrpcPort)
}

func NewConfig() *Config {
	return &Config{GrpcPort: DefaultGrpcPort}
}
