package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/adrg/xdg"
	"github.com/lx7/devnet/internal/auth"
	"github.com/lx7/devnet/internal/client"
	"github.com/lx7/devnet/internal/gui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	conf "github.com/spf13/viper"
)

const (
	appName = "devnet"
	appID   = "net.echocluster.devnet"
)

var g *gui.GUI

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
	runtime.LockOSThread()
}

func configure(confpath string) {
	flag.StringP("config", "c", confpath, "Path to config file")
	flag.StringP("auth.user", "u", "", "Username")
	flag.StringP("auth.pass", "p", "", "Password")
	flag.StringP("log.level", "l", "info", "Loglevel")
	flag.Parse()
	conf.BindPFlags(flag.CommandLine)

	conf.SetConfigFile(conf.GetString("config"))
	if err := conf.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("failed to read config file")
	}

	if ll, err := zerolog.ParseLevel(conf.GetString("log.level")); err != nil {
		log.Error().Err(err).Msg("failed to set log level")
	} else {
		zerolog.SetGlobalLevel(ll)
		if ll >= zerolog.InfoLevel {
			os.Setenv("GST_DEBUG", "*:2")
		}
	}
}

func run() int {
	u := conf.GetString("auth.user")
	p := conf.GetString("auth.pass")
	header := auth.BasicAuthHeader(u, p)

	signal, err := client.Dial(conf.GetString("signaling.URL"), header)
	if err != nil {
		log.Fatal().Err(err).Msg("dial")
	}

	sChan := make(chan *client.Session, 1)
	go func() {
		session, err := client.NewSession(u, signal)
		if err != nil {
			log.Fatal().Err(err).Msg("session")
		}
		sChan <- session
		session.Run()
	}()

	g, err = gui.New(appID, <-sChan)
	if err != nil {
		log.Fatal().Err(err).Msg("gtk app")
	}
	return g.Run()
}

func quit() {
	g.Quit()
}

func main() {
	configure(fmt.Sprintf("%s/%s/config.yaml", xdg.ConfigHome, appName))
	os.Exit(run())
}
