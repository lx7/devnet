package client

import (
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/lx7/devnet/internal/testutil"
	"github.com/lx7/devnet/proto"
	"github.com/pion/webrtc/v2"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type checkFunc func(c *Controller) error
type transitionFunc func(c *Controller) error

func TestController_States(t *testing.T) {
	username := "testuser"
	hook := testutil.NewLogHook()

	webrtcConf := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	signal := &fakeSignal{
		recv: make(chan *proto.Frame, 1),
		send: make(chan *proto.Frame, 1),
	}

	//c.Run()

	// define cases
	tests := []struct {
		desc       string
		giveState  stateFunc
		wantState  stateFunc
		check      checkFunc
		transition transitionFunc
	}{
		{
			desc:      "starting -> idle",
			giveState: stateStarting,
			transition: func(c *Controller) error {
				c.events <- evInitialized
				return nil
			},
			check: func(*Controller) error {
				return nil
			},
			wantState: stateIdle,
		},
		{
			desc:      "idle -> calling",
			giveState: stateIdle,
			transition: func(c *Controller) error {
				return c.AddPeer("user2")
			},
			check: func(*Controller) error {
				return nil
			},
			wantState: stateCalling,
		},
		/*
			{
				desc:      "idle -> calling",
				giveState: stateIdle,
				transition: func(c *Controller) {
					signal.recv <- &proto.SDPMessage{
						Src: "sender",
						Dst: username,
						SDP: webrtc.SessionDescription{
							Type: webrtc.SDPTypeOffer,
							SDP:  "sdp",
						},
					}
				},
				check: func(*Controller) error {
					return nil
				},
				wantState: stateConnected,
			},
		*/
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			c, err := NewController(signal, username, webrtcConf)
			require.NoError(t, err, "constructor should not return error")

			var state stateFunc
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				state = tt.giveState(c)
			}()

			time.Sleep(10 * time.Millisecond)
			require.NoError(t, tt.transition(c), "transition should not return error")

			wg.Wait()
			require.Equal(t, funcName(tt.wantState), funcName(state))
			require.NoError(t, tt.check(c))

			// check for error log entries
			errorlog := hook.Entry(log.ErrorLevel)
			//assert.Equal(t, errorlog, nil)
			if errorlog != nil {
				t.Errorf("runtime error: %v", errorlog.Message)
			}
			hook.Reset()
		})
	}
}

func TestController_EventHandler(t *testing.T) {
	c, err := NewController(nil, "testuser", webrtc.Configuration{})
	require.NoError(t, err)

	h := &fakeEventHandler{}
	h.On("handlerFunc").Return()

	c.Handle(evUndefined, h.handlerFunc)
	c.notify(evUndefined)
	h.AssertExpectations(t)
}

func funcName(f stateFunc) string {
	p := reflect.ValueOf(f).Pointer()
	return runtime.FuncForPC(p).Name()
}

type fakeEventHandler struct {
	mock.Mock
}

func (h *fakeEventHandler) handlerFunc() {
	h.Called()
}

type fakeSignal struct {
	mock.Mock
	recv chan *proto.Frame
	send chan *proto.Frame
}

func (s *fakeSignal) Send(*proto.Frame) error {
	return nil
}

func (s *fakeSignal) Receive() <-chan *proto.Frame {
	return s.recv
}
