package client

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gotk3/gotk3/gtk"
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

	runtime.LockOSThread()
}

func TestSession(t *testing.T) {
	hook := &testutil.LogHook{}
	log.Logger = log.Hook(hook)
	gtk.Init(nil)

	go func() {
		signal1 := &fakeSignal{
			recv: make(chan *proto.Frame, 1),
		}
		signal2 := &fakeSignal{
			recv: make(chan *proto.Frame, 1),
		}
		signal1.other = signal2
		signal2.other = signal1

		s1, err := NewSession("user1", signal1)
		assert.NoError(t, err)
		go s1.Run()

		s2, err := NewSession("user2", signal2)
		assert.NoError(t, err)
		go s2.Run()

		signal1.statehandler(SignalStateConnected)
		signal2.statehandler(SignalStateConnected)

		require.IsType(t, EventConnected{}, <-s1.Events())
		require.IsType(t, EventConnected{}, <-s2.Events())

		// configure
		conf := &proto.Frame{
			Payload: &proto.Frame_Config{Config: &proto.Config{
				Webrtc: &proto.Config_WebRTC{
					Iceservers: []*proto.Config_WebRTC_ICEServer{
						&proto.Config_WebRTC_ICEServer{
							Url: "stun:localhost:19302",
						},
					},
				},
			}},
		}
		signal1.recv <- conf
		signal2.recv <- conf
		time.Sleep(100 * time.Millisecond)

		// test webrtc connection
		err = s1.Connect("user2")
		require.NoError(t, err)

		// get Peer instance from event
		require.IsType(t, EventPeerConnected{}, <-s1.Events())
		ev := <-s2.Events()
		peer1 := ev.(EventPeerConnected).Peer
		assert.Equal(t, "user1", peer1.Name())

		time.Sleep(100 * time.Millisecond)

		// start a stream and check events
		peer1.ScreenLocal().Send()
		select {
		case ev := <-s1.Events():
			assert.IsType(t, ev, EventStreamStart{})
		case <-time.After(1 * time.Second):
			t.Error("receive timeout")
		}

		time.Sleep(2 * time.Second)

		s1.Close()
		s2.Close()
		time.Sleep(10 * time.Millisecond)
		gtk.MainQuit()
	}()

	gtk.Main()

	entry := hook.Entry(zerolog.ErrorLevel)
	if entry != nil {
		log.Fatal().Str("err", entry.Msg).Msg("runtime error")
	}
}

type fakeSignal struct {
	other        *fakeSignal
	recv         chan *proto.Frame
	statehandler func(SignalState)
}

func (s *fakeSignal) Send(f *proto.Frame) error {
	s.other.recv <- f
	return nil
}

func (s *fakeSignal) Receive() <-chan *proto.Frame {
	return s.recv
}

func (s *fakeSignal) Close() error {
	return nil
}

func (s *fakeSignal) HandleStateChange(h SignalStateHandler) {
	s.statehandler = h
}
