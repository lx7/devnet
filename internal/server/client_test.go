package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lx7/devnet/internal/testutil"
	"github.com/lx7/devnet/proto"
	"github.com/pion/webrtc/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Echo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(testutil.Echo))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	time.Sleep(100 * time.Millisecond)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)

	sw := &fakeSwitch{
		forward: make(chan *proto.SDPMessage),
	}

	client := NewClient(conn, "client 1")
	assert.Equal(t, client.Name(), "client 1", "client name shpuld match")
	sw.Register(client)

	// test echo
	give := &proto.SDPMessage{
		Src: "user 1",
		Dst: "user 2",
		SDP: webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  "sdp",
		},
	}
	client.Send() <- give
	time.Sleep(100 * time.Millisecond)

	select {
	case res := <-sw.forward:
		assert.Equal(t, give, res, "response should be equal to message")
	case <-time.After(1 * time.Second):
		t.Error("receive timeout")
	}

	server.Close()
}

type fakeSwitch struct {
	client  Client
	forward chan *proto.SDPMessage
}

func (s *fakeSwitch) Register(c Client) {
	s.client = c
	go c.Attach(s)
}

func (s *fakeSwitch) Forward() chan<- *proto.SDPMessage {
	return s.forward
}

func (*fakeSwitch) Run() {
}

func (*fakeSwitch) Shutdown() {
}
