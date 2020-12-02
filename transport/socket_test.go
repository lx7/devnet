package transport

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestSocket_NotConnected(t *testing.T) {
	socket := &Socket{}
	defer socket.Close()

	// define cases
	tests := []struct {
		desc    string
		give    func() error
		wantErr error
	}{
		{
			desc: "WriteMessage",
			give: func() error {
				return socket.WriteMessage([]byte("some message"))
			},
			wantErr: ErrWSNotConnected,
		},
		{
			desc: "ReadMessage",
			give: func() error {
				_, err := socket.ReadMessage()
				return err
			},
			wantErr: ErrWSNotConnected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			err := tt.give()
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestSocket_EchoReconnect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(echo))
	url := "ws" + strings.TrimPrefix(server.URL, "http")

	socket := Dial(url, nil)
	defer socket.Close()
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, true, socket.Connected(), "socket should be connected")

	msg := []byte("-- { message } //")

	// test echo
	require.NoError(t, socket.WriteMessage(msg))
	resp, err := socket.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, string(resp), string(msg), "response mismatch")

	// force reconnect
	_ = socket.ws.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	time.Sleep(100 * time.Millisecond)
	log.SetOutput(ioutil.Discard)
	require.Error(t, socket.WriteMessage(msg))
	log.SetOutput(os.Stdout)

	// test echo again
	time.Sleep(5 * time.Second)
	require.NoError(t, socket.WriteMessage(msg))

	resp, err = socket.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, string(resp), string(msg), "response mismatch")

	server.Close()
}

func echo(w http.ResponseWriter, r *http.Request) {
	socket, err := Upgrade(w, r)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer socket.Close()

	for {
		message, err := socket.ReadMessage()
		if err != nil {
			return
		}
		if err := socket.WriteMessage(message); err != nil {
			return
		}
	}
}
