package main

import (
	"fmt"
	"os"
	"strings"

	"path/filepath"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var (
	Version   string = "DEV"
	config           = NewConfig()
	templates *template.Template
)

func ReadTemplates() {
	pattern := filepath.Join(config.TemplatesPath, "*.tmpl")
	templates = template.Must(template.ParseGlob(pattern))
}

func RunAction(c *cli.Context) {
	config.Debug = c.Bool("debug")
	config.GrpcPort = c.Int("grpc-port")
	config.TlsCertFile = c.String("tls-cert-file")
	config.TlsKeyFile = c.String("tls-key-file")
	config.ApiServiceURL = c.String("api-service")
	config.ApiConnectMode = c.String("api-connect-mode")
	config.TemplatesPath = c.String("templates-path")

	if config.Debug {
		log.SetLevel(log.DebugLevel)
	}

	config.ReadFiles()
	ReadTemplates()

	server := NewServer(config)
	server.Listen()
}

func withEnvs(prefix string, flags []cli.Flag) []cli.Flag {

	var flgs []cli.Flag
	for _, f := range flags {
		env := ""
		spr := strings.Split(f.GetName(), ",")
		env = prefix + "_" + strings.ToUpper(strings.Replace(spr[0], "-", "_", -1))
		switch v := f.(type) {
		case cli.IntFlag:
			flgs = append(flgs, cli.IntFlag{Name: v.Name, Value: v.Value, Usage: v.Usage, EnvVar: env})
		case cli.StringFlag:
			flgs = append(flgs, cli.StringFlag{Name: v.Name, Value: v.Value, Usage: v.Usage, EnvVar: env})
		case cli.StringSliceFlag:
			flgs = append(flgs, cli.StringSliceFlag{Name: v.Name, Value: v.Value, Usage: v.Usage, EnvVar: env})
		case cli.BoolFlag:
			flgs = append(flgs, cli.BoolFlag{Name: v.Name, Usage: v.Usage, EnvVar: env})
		default:
			fmt.Println("unknown")
		}
	}
	return flgs
}

func main() {
	app := cli.NewApp()
	app.Name = "welcome"
	app.Version = Version
	app.Usage = "Sample Dashboard Data Provider, Only sends welcome card if user is registered in last one week"
	app.Author = "Sercan Degirmenci <sercan@otsimo.com>"

	flags := []cli.Flag{
		cli.IntFlag{Name: "grpc-port", Value: config.GrpcPort, Usage: "grpc server port"},
		cli.StringFlag{Name: "tls-cert-file", Value: "", Usage: "the server's certificate file for TLS connection"},
		cli.StringFlag{Name: "tls-key-file", Value: "", Usage: "the server's private key file for TLS connection"},
		cli.StringFlag{Name: "client-id", Value: "", Usage: "client id"},
		cli.StringFlag{Name: "client-secret", Value: "", Usage: "client secret"},
		cli.StringFlag{Name: "discovery", Value: "https://connect.otsimo.com", Usage: "auth discovery url"},
		cli.StringFlag{Name: "api-service", Value: "api.default", Usage: "api service url"},
		cli.StringFlag{Name: "api-connect-mode", Value: "insecure-tls", Usage: "connection credentials mode, valid values are: tls, insecure-tls, insecure"},
		cli.StringFlag{Name: "templates-path", Value: "./templates", Usage: "the folder that contains templates"},

		cli.BoolFlag{Name: "debug, d", Usage: "enable verbose log"},
	}
	flags = withEnvs("OTSIMO_WELCOME", flags)

	app.Flags = flags
	app.Action = RunAction

	log.Infoln("running", app.Name, "version:", app.Version)
	app.Run(os.Args)
}

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}
