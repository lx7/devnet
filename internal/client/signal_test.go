package client

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lx7/devnet/internal/testutil"
	"github.com/lx7/devnet/proto"
	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func TestSignal_Echo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(testutil.Echo))
	url := "ws" + strings.TrimPrefix(server.URL, "http")

	time.Sleep(100 * time.Millisecond)
	signal, err := Dial(url, nil)
	assert.NoError(t, err)

	// test echo
	give := &proto.SDPMessage{
		Src: "user 1",
		Dst: "user 2",
		SDP: webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  "sdp",
		},
	}
	signal.Send(give)
	time.Sleep(100 * time.Millisecond)

	select {
	case res := <-signal.Receive():
		assert.Equal(t, give, res, "response should be equal to message")
	case <-time.After(1 * time.Second):
		t.Error("receive timeout")
	}

	time.Sleep(100 * time.Millisecond)
	signal.Close()
	// wait for connection timeout
	time.Sleep(600 * time.Millisecond)

	server.Close()
}
