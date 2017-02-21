package main

import (
	"dashboard"
	"dashboard/storage"
	_ "dashboard/storage/mongodb"
	_ "dashboard/storage/postgres"
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"google.golang.org/grpc/grpclog"
)

var Version string = "DEV"
var config = dashboard.NewConfig()

func RunAction(c *cli.Context) error {
	if config.Debug {
		log.SetLevel(log.DebugLevel)
	}
	//find driver name
	sname := c.String("storage")
	if sname == "" || sname == "none" {
		cli.ShowAppHelp(c)
		return fmt.Errorf("main.go: storage flag='%s' is invalid", sname)
	}

	//get driver
	driver := storage.GetDriver(sname)
	if driver == nil {
		return fmt.Errorf("main.go: storage driver '%s' not found\n", sname)
	}

	//load storage driver
	s, err := driver.New(c)
	if err != nil {
		return fmt.Errorf("main.go: error while creating new storage[%s] driver: %v", s, err)
	}

	server := dashboard.NewServer(config, s)
	return server.Listen()
}

func withEnvs(prefix string, flags []cli.Flag) []cli.Flag {
	var flgs []cli.Flag
	for _, f := range flags {
		env := ""
		spr := strings.Split(f.GetName(), ",")
		env = prefix + "_" + strings.ToUpper(strings.Replace(spr[0], "-", "_", -1))
		switch v := f.(type) {
		case cli.IntFlag:
			flgs = append(flgs, cli.IntFlag{Name: v.Name, Destination: v.Destination, Value: v.Value, Usage: v.Usage, EnvVar: env})
		case cli.StringFlag:
			flgs = append(flgs, cli.StringFlag{Name: v.Name, Destination: v.Destination, Value: v.Value, Usage: v.Usage, EnvVar: env})
		case cli.BoolFlag:
			flgs = append(flgs, cli.BoolFlag{Name: v.Name, Destination: v.Destination, Usage: v.Usage, EnvVar: env})
		default:
			fmt.Println("unknown")
		}
	}
	return flgs
}

func main() {
	app := cli.NewApp()
	app.Name = "otsimo-dashboard"
	app.Version = Version
	app.Usage = "Otsimo Dashboard Service"
	app.Author = "Sercan Degirmenci <sercan@otsimo.com>"
	dnames := storage.GetDriverNames()
	var flags []cli.Flag

	flags = []cli.Flag{
		cli.IntFlag{Name: "grpc-port", Destination: &config.GrpcPort, Value: config.GrpcPort, Usage: "grpc server port"},
		cli.IntFlag{Name: "health-port", Destination: &config.HealthPort, Value: config.HealthPort, Usage: "health check server port"},
		cli.StringFlag{Name: "storage, s", Value: "none", Usage: fmt.Sprintf("the storage driver. Available drivers: %s", strings.Join(dnames, ", "))},
		cli.StringFlag{Name: "tls-cert-file", Destination: &config.TlsCertFile, Usage: "the server's certificate file for TLS connection"},
		cli.StringFlag{Name: "tls-key-file", Destination: &config.TlsKeyFile, Usage: "the server's private key file for TLS connection"},
		cli.StringFlag{Name: "client-id", Destination: &config.ClientID, Usage: "client id"},
		cli.StringFlag{Name: "client-secret", Destination: &config.ClientSecret, Usage: "client secret"},
		cli.StringFlag{Name: "discovery", Destination: &config.AuthDiscovery, Value: config.AuthDiscovery, Usage: "auth discovery url"},
		cli.StringFlag{Name: "config-path", Destination: &config.ConfigPath, Value: config.ConfigPath, Usage: "config file path"},
		cli.StringFlag{Name: "default-lang", Destination: &config.DefaultLanguage, Usage: "default language"},
		cli.BoolFlag{Name: "debug, d", Destination: &config.Debug, Usage: "enable verbose log"},
		cli.BoolFlag{Name: "watch-config", Destination: &config.WatchConfigFile, Usage: "watch configuration file for changes"},
		cli.BoolFlag{Name: "no-auth", Destination: &config.NoAuth, Usage: "do not try to get an access token"},
	}
	flags = withEnvs("OTSIMO_DASHBOARD", flags)
	for _, d := range dnames {
		flags = append(flags, storage.GetDriver(d).Flags...)
	}
	app.Flags = flags
	app.Action = RunAction

	log.Infoln("running", app.Name, "version:", app.Version)
	if err := app.Run(os.Args); err != nil {
		log.Errorf("failed to run app %v", err)
	}
}

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableColors: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	grpclog.SetLogger(&log.Logger{
		Out:       os.Stdout,
		Formatter: &log.TextFormatter{FullTimestamp: true, DisableColors: true},
		Hooks:     make(log.LevelHooks),
		Level:     log.GetLevel(),
	})
}
