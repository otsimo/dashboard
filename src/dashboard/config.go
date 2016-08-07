package dashboard

import "fmt"

const (
	DefaultGrpcPort   = 18860
	DefaultHealthPort = 8080
)

type CommandConfig struct {
	Debug           bool
	GrpcPort        int
	HealthPort      int
	TlsCertFile     string
	TlsKeyFile      string
	ClientID        string
	ClientSecret    string
	TrustedCAFile   string
	AuthDiscovery   string
	ConfigPath      string
	WatchConfigFile bool
	NoAuth          bool
	DefaultLanguage string
}

func (c *CommandConfig) GetGrpcPortString() string {
	return fmt.Sprintf(":%d", c.GrpcPort)
}

func (c *CommandConfig) GetHealthPortString() string {
	return fmt.Sprintf(":%d", c.HealthPort)
}

func NewConfig() *CommandConfig {
	return &CommandConfig{
		GrpcPort:      DefaultGrpcPort,
		HealthPort:    DefaultHealthPort,
		ConfigPath:    "config.yaml",
		AuthDiscovery: "https://connect.otsimo.com",
	}
}

type ServiceConfig struct {
	Providers []ProviderConfig `json:"providers"`
}
