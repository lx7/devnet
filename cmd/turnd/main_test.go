package main

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/lx7/devnet/internal/testutil"
	"github.com/pion/turn/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	conf "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
	configure("../../configs/turnd.yaml")
}

func TestTurnD_Config(t *testing.T) {
	// flag defaults
	assert.Equal(t, "info", conf.GetString("loglevel"))

	// values from config file
	assert.Equal(t, "0.0.0.0", conf.GetString("turn.ip"))
	assert.Equal(t, "3478", conf.GetString("turn.port"))
	assert.Equal(t, "devnet.test", conf.GetString("turn.realm"))
}

func TestTurnD_STUN(t *testing.T) {
	// initialize log hook
	hook := &testutil.LogHook{}
	log.Logger = log.Hook(hook)

	// run the server and allow for some startup time
	go run()
	time.Sleep(10 * time.Millisecond)

	t.Run("stun binding request", func(t *testing.T) {
		conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
		assert.NoError(t, err)

		c, err := turn.NewClient(&turn.ClientConfig{
			Conn: conn,
		})
		assert.NoError(t, err)
		assert.NoError(t, c.Listen())

		_, err = c.SendBindingRequestTo(&net.UDPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 3478,
		})
		assert.NoError(t, err, "stun request should succeed")

		c.Close()
		assert.NoError(t, conn.Close())
	})

	// check the log for errors
	entry := hook.Entry(zerolog.ErrorLevel)
	if entry != nil {
		log.Fatal().Str("err", entry.Msg).Msg("runtime error")
	}
}
