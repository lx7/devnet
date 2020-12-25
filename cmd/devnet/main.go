package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/adrg/xdg"
	"github.com/lx7/devnet/internal/auth"
	"github.com/lx7/devnet/internal/client"
	"github.com/lx7/devnet/internal/gui"
	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
	conf "github.com/spf13/viper"
)

const (
	appName = "devnet"
	appID   = "net.echocluster.devnet"
)

func init() {
	runtime.LockOSThread()
	os.Setenv("GST_DEBUG", "*:2")
}

func configure(confpath string) {
	conf.SetDefault("config", confpath)
	conf.SetDefault("loglevel", "info")

	conf.BindEnv("loglevel", "DEVNET_LOGLEVEL")
	conf.BindEnv("config", "DEVNET_CONFIG")
	conf.BindEnv("user.name", "DEVNET_USER")
	conf.BindEnv("user.pass", "DEVNET_PASS")

	conf.SetConfigFile(conf.GetString("config"))
	if err := conf.ReadInConfig(); err != nil {
		log.Fatal("failed reading config file: ", err)
	}

	loglevel, err := log.ParseLevel(conf.GetString("loglevel"))
	if err != nil {
		log.Error("failed to set log level: ", err)
	}
	log.SetLevel(loglevel)
}

func run() {
	header := auth.BasicAuthHeader(
		conf.GetString("user.name"),
		conf.GetString("user.pass"),
	)
	signal, err := client.Dial(conf.GetString("signaling.URL"), header)
	if err != nil {
		log.Fatal("dial: ", err)
	}

	wconf := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{
			URLs: []string{conf.GetString("stun.URL")},
		}},
	}

	sChan := make(chan *client.Session, 1)
	go func() {
		session, err := client.NewSession(signal, client.SessionOpts{
			Self:       conf.GetString("user.name"),
			WebRTCConf: wconf,
		})
		if err != nil {
			log.Fatal("client controller: ", err)
		}
		sChan <- session
		session.Run()
	}()

	g, err := gui.New(appID, <-sChan)
	if err != nil {
		log.Fatal(err)
	}
	os.Exit(g.Run())
}

func main() {
	configure(fmt.Sprintf("%s/%s/config.yaml", xdg.ConfigHome, appName))
	run()
}
