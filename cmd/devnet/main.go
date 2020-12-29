package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
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

var _gui *gui.GUI

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
		FormatFieldValue: func(i interface{}) string {
			s := fmt.Sprintf("%s", i)
			s = strings.Replace(s, `\n`, "\n", -1)
			s = strings.Replace(s, `\t`, "\t", -1)
			return s
		},
	})
	runtime.LockOSThread()
}

func configure(confpath string) {
	flag.StringP("config", "c", confpath, "config file")
	flag.StringP("loglevel", "l", "info", "log level")
	flag.Parse()
	conf.RegisterAlias("log.level", "loglevel")
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
	url := conf.GetString("signaling.URL")

	header := auth.BasicAuthHeader(u, p)
	signal := client.Dial(url, header)

	sChan := make(chan *client.Session, 1)
	go func() {
		session, err := client.NewSession(u, signal)
		if err != nil {
			log.Fatal().Err(err).Msg("session")
		}
		sChan <- session
		session.Run()
	}()

	g, err := gui.New(appID, <-sChan)
	if err != nil {
		log.Fatal().Err(err).Msg("gtk app")
	}
	_gui = g
	return _gui.Run()
}

func main() {
	configure(fmt.Sprintf("%s/%s/config.yaml", xdg.ConfigHome, appName))
	os.Exit(run())
}
