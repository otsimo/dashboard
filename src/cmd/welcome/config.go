package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
)

const (
	DefaultGrpcPort = 18864
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
	ApiServiceURL string
}

func (c *Config) GetGrpcPortString() string {
	return fmt.Sprintf(":%d", c.GrpcPort)
}

func (c *Config) ReadFiles() {
	_, err := os.Stat(c.ClientID)
	if !os.IsNotExist(err) {
		d, err := ioutil.ReadFile(c.ClientID)
		if err == nil {
			c.ClientID = string(d)
		} else {
			logrus.Errorf("config.go: failed to read clientID=%s err=%v", c.ClientID, err)
		}
	}

	_, err = os.Stat(c.ClientSecret)
	if !os.IsNotExist(err) {
		d, err := ioutil.ReadFile(c.ClientSecret)
		if err == nil {
			c.ClientSecret = string(d)
		} else {
			logrus.Errorf("config.go: failed to read err=%v", err)
		}
	}
}

func NewConfig() *Config {
	return &Config{GrpcPort: DefaultGrpcPort}
}
