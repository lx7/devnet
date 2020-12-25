package client

import (
	"runtime"
	"testing"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/internal/testutil"
	"github.com/lx7/devnet/proto"
	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.LockOSThread()
	//os.Setenv("GST_DEBUG", "*:2")
	//log.SetLevel(log.DebugLevel)
}

func TestSession_Flow(t *testing.T) {
	hook := testutil.NewLogHook()
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

		conf := webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{{URLs: []string{
				"stun:stun.l.google.com:19302",
			}}},
		}

		s1, err := NewSession(signal1, SessionOpts{
			Self:       "user1",
			WebRTCConf: conf,
		})
		require.NoError(t, err)
		go s1.Run()

		s2, err := NewSession(signal2, SessionOpts{
			Self:       "user2",
			WebRTCConf: conf,
		})
		require.NoError(t, err)
		go s2.Run()

		err = s1.Connect("user2")
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		s1.StartStream(LocalScreen)

		time.Sleep(5 * time.Second)
		s1.Close()
		//s2.Close()
		time.Sleep(50 * time.Millisecond)
		gtk.MainQuit()
	}()

	gtk.Main()

	errorlog := hook.Entry(log.ErrorLevel)
	if errorlog != nil {
		t.Errorf("runtime error: '%v'", errorlog.Message)
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
