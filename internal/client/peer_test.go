package client

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lx7/devnet/internal/testutil"
	"github.com/lx7/devnet/proto"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func TestPeer(t *testing.T) {
	hook := &testutil.LogHook{}
	log.Logger = log.Hook(hook)

	events := make(chan Event, 5)
	c1 := make(chan *proto.Frame, 5)
	c2 := make(chan *proto.Frame, 5)
	done := make(chan bool)

	p1, err := NewPeer("peer1", c1, events)
	require.NoError(t, err)
	go p1.Run()

	p2, err := NewPeer("peer2", c2, events)
	require.NoError(t, err)
	go p2.Run()

	go func() {
		for {
			select {
			case frame := <-c1:
				frame.Src = "peer2"
				err := p2.HandleSignaling(frame)
				require.NoError(t, err)
			case frame := <-c2:
				frame.Src = "peer1"
				err := p1.HandleSignaling(frame)
				require.NoError(t, err)
			case <-done:
				return
			}
		}
	}()

	// test webrtc connection
	err = p1.Connect()
	require.NoError(t, err)

	select {
	case have := <-events:
		assert.IsType(t, EventPeerConnected{}, have)
		return
	case <-time.After(1 * time.Second):
		t.Error("receive timeout")
	}

	time.Sleep(100 * time.Millisecond)

	// send remote control message
	give := &proto.Control{
		Time: 100,
	}
	p2.RCon(give)

	select {
	case have := <-events:
		assert.IsType(t, EventRCon{}, have)
	case <-time.After(1 * time.Second):
		t.Error("receive timeout")
	}

	time.Sleep(2 * time.Second)

	close(done)
	p1.Close()
	p2.Close()
	time.Sleep(10 * time.Millisecond)

	entry := hook.Entry(zerolog.ErrorLevel)
	if entry != nil {
		log.Fatal().Str("err", entry.Msg).Msg("runtime error")
	}
}
