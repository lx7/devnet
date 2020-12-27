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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "google.golang.org/protobuf/proto"
)

func TestClient_Echo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(testutil.Echo))
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	time.Sleep(100 * time.Millisecond)

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)

	sw := &fakeSwitch{
		forward: make(chan *proto.Frame),
	}

	client := NewClient(conn, "client 1")
	assert.Equal(t, client.Name(), "client 1", "client name should match")
	sw.Register(client)

	give := &proto.Frame{
		Src: "user 1",
		Dst: "user 2",
	}
	client.Send() <- give
	time.Sleep(100 * time.Millisecond)

	select {
	case have := <-sw.forward:
		assert.True(t, pb.Equal(give, have), "response should match")
	case <-time.After(1 * time.Second):
		t.Error("receive timeout")
	}

	server.Close()
}

type fakeSwitch struct {
	client  Client
	forward chan *proto.Frame
}

func (s *fakeSwitch) Register(c Client) {
	s.client = c
	go c.Attach(s)
}

func (s *fakeSwitch) Unregister(c Client) {
	s.client = nil
}

func (s *fakeSwitch) Forward() chan<- *proto.Frame {
	return s.forward
}

func (*fakeSwitch) Run() {
}

func (*fakeSwitch) Shutdown() {
}
