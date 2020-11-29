package transport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lx7/devnet/proto"
	"github.com/pion/webrtc/v2"
)

func init() {
	//log.SetLevel(log.ErrorLevel)
}

func TestMux(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	socket := Dial(url, nil)
	time.Sleep(100 * time.Millisecond)
	defer socket.Close()

	c := &testConsumer{
		SDPReceived: make(chan bool),
	}
	mux := NewMux(socket, c)
	go func() {
		err := mux.Receive()
		if err != nil {
			t.Error(err)
		}
	}()

	msg := &proto.SDPMessage{
		Src: "",
		Dst: "",
		SDP: webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  "sdp",
		},
	}
	mux.Send(msg)

	select {
	case ok := <-c.SDPReceived:
		if !ok {
			t.Errorf("sdp message delivery")
		}
		socket.Close()
		time.Sleep(100 * time.Millisecond)
	case <-time.After(100 * time.Millisecond):
		t.Errorf("sdp message timeout")
	}
}

type testConsumer struct {
	SDPReceived chan bool
}

func (c *testConsumer) HandleSDP(m *proto.SDPMessage) {
	c.SDPReceived <- true
}
