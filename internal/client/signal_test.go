package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lx7/devnet/internal/testutil"
	"github.com/lx7/devnet/proto"

	"github.com/stretchr/testify/assert"
)

func TestSignal_Echo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(testutil.Echo))
	url := "ws" + strings.TrimPrefix(server.URL, "http")

	time.Sleep(100 * time.Millisecond)
	signal := Dial(url, nil)

	// test echo
	give := &proto.Frame{
		Src: "user 1",
		Dst: "user 2",
		Payload: &proto.Frame_Sdp{Sdp: &proto.SDP{
			Type: proto.SDP_OFFER,
			Desc: "sdp",
		}},
	}
	signal.Send(give)
	time.Sleep(100 * time.Millisecond)

	select {
	case have := <-signal.Receive():
		assert.IsType(t, give, have, "response should match")
	case <-time.After(1 * time.Second):
		t.Error("receive timeout")
	}

	time.Sleep(100 * time.Millisecond)
	signal.Close()
	time.Sleep(100 * time.Millisecond)
	server.Close()
}
