package main

import (
	"fmt"

	"github.com/lx7/devnet/internal/server"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	conf "github.com/spf13/viper"
)

const appName = "devnet"

func init() {
	log.SetFormatter(&log.TextFormatter{
		//FullTimestamp: true,
	})
}

func configure(confpath string) {
	flag.StringP("loglevel", "l", "info", "Log level")
	flag.StringP("config", "c", confpath, "Path to config file")
	flag.Parse()
	conf.BindPFlags(flag.CommandLine)

	conf.SetConfigFile(conf.GetString("config"))
	err := conf.ReadInConfig()
	if err != nil {
		log.Fatal("failed reading config file: ", err)
	}

	if loglevel, err := log.ParseLevel(conf.GetString("loglevel")); err != nil {
		log.Error("failed to set log level: ", err)
	} else {
		log.SetLevel(loglevel)
	}
}

func run() {
	s := server.New(conf.GetString("server.addr"))
	s.Serve(conf.GetString("server.wspath"))
}

func main() {
	configure(fmt.Sprintf("/etc/%s/config.yaml", appName))
	run()
}
