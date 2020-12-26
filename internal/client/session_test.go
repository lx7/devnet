package client

import (
	"runtime"
	"testing"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/internal/testutil"
	"github.com/lx7/devnet/proto"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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

		s1, err := NewSession("user1", signal1)
		assert.NoError(t, err)
		go s1.Run()

		s2, err := NewSession("user2", signal2)
		assert.NoError(t, err)
		go s2.Run()

		// configure
		conf := &proto.Frame{
			Dst: "user1",
			Payload: &proto.Frame_Config{&proto.Config{
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
