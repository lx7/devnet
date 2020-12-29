package main

import (
	"fmt"
	"os"
	"time"

	"github.com/lx7/devnet/internal/server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	conf "github.com/spf13/viper"
)

const appName = "devnet"

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
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
		log.Fatal().Err(err).Msg("failed to read config file")
	}

	if ll, err := zerolog.ParseLevel(conf.GetString("loglevel")); err != nil {
		log.Error().Err(err).Msg("failed to set log level")
	} else {
		zerolog.SetGlobalLevel(ll)
	}
}

func run() {
	s := server.New(conf.GetViper())
	err := s.Serve()
	if err != nil {
		log.Fatal().Err(err).Msg("http server")
	}
}

func main() {
	configure(fmt.Sprintf("/etc/%s/signald.yaml", appName))
	run()
}
