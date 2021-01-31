package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lx7/devnet/internal/auth"
	"github.com/lx7/devnet/internal/turn"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	conf "github.com/spf13/viper"
)

const appName = "devnet"

func init() {
	/*
			syslog, err := syslog.New(syslog.LOG_WARNING|syslog.LOG_DAEMON, "")
			if err != nil {
				log.Fatal().Err(err).Msg("syslog")
			}
			log.Logger = log.Output(zerolog.SyslogLevelWriter(syslog))
		}
	*/
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

	err = auth.Configure(conf.Sub("auth"))
	if err != nil {
		log.Fatal().Err(err).Msg("configure auth")
	}
}

func run() {
	ip := net.ParseIP(conf.GetString("turn.ip"))
	port := conf.GetInt("turn.port")
	realm := conf.GetString("turn.realm")

	if ip == nil {
		log.Fatal().Msgf("ip address not configured")
	} else if port == 0 {
		log.Fatal().Msgf("port not configured")
	}

	s, err := turn.NewServer(ip, port, realm)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start turn server")
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	if err = s.Close(); err != nil {
		log.Fatal().Err(err).Msg("failed to close turn server")
	}
}

func main() {
	configure("/etc/devnet/turnd.yaml")
	run()
}
