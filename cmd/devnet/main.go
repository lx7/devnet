package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/adrg/xdg"
	"github.com/lx7/devnet/internal/auth"
	"github.com/lx7/devnet/internal/client"
	"github.com/lx7/devnet/internal/gui"
	log "github.com/sirupsen/logrus"
	conf "github.com/spf13/viper"
)

const (
	appName = "devnet"
	appID   = "net.echocluster.devnet"
)

var g *gui.GUI

func init() {
	runtime.LockOSThread()
}

func configure(confpath string) {
	conf.SetDefault("config", confpath)
	conf.SetDefault("log.level", "info")
	conf.BindEnv("config", "DEVNET_CONFIG")

	conf.SetConfigFile(conf.GetString("config"))
	if err := conf.ReadInConfig(); err != nil {
		log.Fatal("failed reading config file: ", err)
	}

	loglevel, err := log.ParseLevel(conf.GetString("log.level"))
	if err != nil {
		log.Error("failed to set log level: ", err)
	}
	log.SetLevel(loglevel)

	if log.GetLevel() >= log.InfoLevel {
		os.Setenv("GST_DEBUG", "*:2")
	}
}

func run() int {
	u := conf.GetString("auth.user")
	p := conf.GetString("auth.pass")
	header := auth.BasicAuthHeader(u, p)

	signal, err := client.Dial(conf.GetString("signaling.URL"), header)
	if err != nil {
		log.Fatal("dial: ", err)
	}

	sChan := make(chan *client.Session, 1)
	go func() {
		session, err := client.NewSession(u, signal)
		if err != nil {
			log.Fatal("client controller: ", err)
		}
		sChan <- session
		session.Run()
	}()

	g, err = gui.New(appID, <-sChan)
	if err != nil {
		log.Fatal(err)
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
