package dashboard

import "fmt"

const (
	DefaultGrpcPort = 18860
)

type CommandConfig struct {
	Debug           bool
	GrpcPort        int
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

func NewConfig() *CommandConfig {
	return &CommandConfig{GrpcPort: DefaultGrpcPort}
}

type ServiceConfig struct {
	Providers []ProviderConfig `json:"providers"`
}
