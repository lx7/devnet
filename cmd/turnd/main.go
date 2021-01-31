package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pion/turn/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	flag "github.com/spf13/pflag"
	conf "github.com/spf13/viper"
)

const appName = "devnet"

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
	ip := conf.GetString("turn.ip")
	port := conf.GetString("turn.port")
	realm := conf.GetString("turn.realm")

	if ip == "" {
		log.Fatal().Msgf("ip address not configured")
	} else if port == "" {
		log.Fatal().Msgf("port not configured")
	}

	listener, err := net.ListenPacket("udp4", "0.0.0.0:"+port)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create udp listener")
	}

	s, err := turn.NewServer(turn.ServerConfig{
		Realm: realm,
		AuthHandler: func(user string, realm string, srcAddr net.Addr) ([]byte, bool) {
			key := turn.GenerateAuthKey(user, realm, "pass")
			if len(key) > 0 {
				return key, false
			}
			return nil, false
		},
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: listener,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(ip),
					Address:      "0.0.0.0",
				},
			},
		},
	})
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
