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

func TestSession_Flow(t *testing.T) {
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

		// configure
		conf := &proto.Frame{
			Dst: "user1",
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
		time.Sleep(10 * time.Millisecond)

		// test webrtc connection
		err = s1.Connect("user2")
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		s1.StartStream(LocalScreen)

		time.Sleep(5 * time.Second)

		// TODO: investigate io errors after close
		//s1.Close()
		//s2.Close()
		time.Sleep(50 * time.Millisecond)
		gtk.MainQuit()
	}()

	gtk.Main()

	entry := hook.Entry(zerolog.ErrorLevel)
	if entry != nil {
		log.Fatal().Str("err", entry.Msg).Msg("runtime error")
	}
}

type fakeSignal struct {
	other *fakeSignal
	recv  chan *proto.Frame
}

func (s *fakeSignal) Send(f *proto.Frame) error {
	s.other.recv <- f
	return nil
}

func (s *fakeSignal) Receive() <-chan *proto.Frame {
	return s.recv
}
