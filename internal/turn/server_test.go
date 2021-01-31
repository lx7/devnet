package turn

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/lx7/devnet/internal/auth"
	"github.com/lx7/devnet/internal/testutil"
	"github.com/pion/turn/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	conf := viper.New()
	conf.SetConfigFile("../../configs/signald.yaml")
	if err := conf.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("read config file")
	}

	err := auth.Configure(conf.Sub("auth"))
	if err != nil {
		log.Fatal().Err(err).Msg("configure")
	}
}

func TestServer(t *testing.T) {
	// initialize log hook
	hook := &testutil.LogHook{}
	log.Logger = log.Hook(hook)

	s, err := NewServer(net.IPv4(127, 0, 0, 1), 3478, "devnet.test")
	assert.NoError(t, err)

	t.Run("stun binding request", func(t *testing.T) {
		listener, err := net.ListenPacket("udp4", "0.0.0.0:0")
		assert.NoError(t, err)

		c, err := turn.NewClient(&turn.ClientConfig{
			Conn: listener,
		})
		assert.NoError(t, err)
		assert.NoError(t, c.Listen())

		_, err = c.SendBindingRequestTo(&net.UDPAddr{
			IP:   net.IPv4(127, 0, 0, 1),
			Port: 3478,
		})
		assert.NoError(t, err)

		c.Close()
		assert.NoError(t, listener.Close())

		// check the log for errors
		entry := hook.Entry(zerolog.ErrorLevel)
		require.Nil(t, entry, "no runtime errors expected")
		hook.Reset()
	})

	t.Run("authentication failure", func(t *testing.T) {
		listener, err := net.ListenPacket("udp4", "0.0.0.0:0")
		assert.NoError(t, err)

		c, err := turn.NewClient(&turn.ClientConfig{
			Conn:           listener,
			TURNServerAddr: "127.0.0.1:3478",
			Username:       "testuser",
			Password:       "wrong password",
		})
		assert.NoError(t, err)
		assert.NoError(t, c.Listen())

		// send a turn alocation request
		_, err = c.Allocate()
		require.EqualError(t, err, "Allocate error response (error 400: )")

		c.Close()
		assert.NoError(t, listener.Close())

		// check the log for errors
		entry := hook.Entry(zerolog.ErrorLevel)
		require.NotNil(t, entry, "runtime error expected")
		hook.Reset()
	})

	t.Run("turn echo", func(t *testing.T) {
		listener, err := net.ListenPacket("udp4", "0.0.0.0:0")
		assert.NoError(t, err)

		c, err := turn.NewClient(&turn.ClientConfig{
			Conn:           listener,
			TURNServerAddr: "127.0.0.1:3478",
			Username:       "testuser",
			Password:       "test",
		})
		assert.NoError(t, err)
		assert.NoError(t, c.Listen())

		// send a turn alocation request
		conn, err := c.Allocate()
		require.NoError(t, err)

		log.Debug().Msgf("laddr: %s", conn.LocalAddr().String())

		c.Close()
		assert.NoError(t, listener.Close())

		// check the log for errors
		entry := hook.Entry(zerolog.ErrorLevel)
		require.Nil(t, entry, "no runtime errors expected")
		hook.Reset()
	})

	assert.NoError(t, s.Close())
}
