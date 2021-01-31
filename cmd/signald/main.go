package main

import (
	"fmt"
	"log/syslog"

	"github.com/lx7/devnet/internal/signaling"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	conf "github.com/spf13/viper"
)

const appName = "devnet"

func init() {
	syslog, err := syslog.New(syslog.LOG_WARNING|syslog.LOG_DAEMON, "")
	if err != nil {
		log.Fatal().Err(err).Msg("syslog")
	}
	log.Logger = log.Output(zerolog.SyslogLevelWriter(syslog))
	/*
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})
	*/
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
	s := signaling.NewServer(conf.GetViper())
	err := s.Serve()
	if err != nil {
		log.Fatal().Err(err).Msg("http server")
	}
}

func main() {
	configure(fmt.Sprintf("/etc/%s/signald.yaml", appName))
	run()
}
