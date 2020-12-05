package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	conf "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLevel(log.InfoLevel)

	conf.Set("users.testuser",
		"09d9623a149a4a0c043befcb448c9c3324be973230188ba412c008a2929f31d0")

	s := New("127.0.0.1:40100")
	go s.Serve("/channel")
	time.Sleep(20 * time.Millisecond)
}

func TestServer_Websocket(t *testing.T) {
	// connect websocket
	header := make(http.Header)
	header.Add("Authorization", "Basic dGVzdHVzZXI6dGVzdA==")
	url := "ws://127.0.0.1:40100/channel"
	ws, _, err := websocket.DefaultDialer.Dial(url, header)
	require.NoError(t, err)
	defer ws.Close()

	// define cases
	tests := []struct {
		desc     string
		giveStr  string
		giveType int
		wantStr  string
	}{
		{
			desc: "sdp echo",
			giveStr: `{
				"type":"sdp", 
				"src":"testuser", 
				"dst":"testuser", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
			giveType: websocket.TextMessage,
			wantStr: `{
				"type":"sdp", 
				"src":"testuser", 
				"dst":"testuser", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
		},
	}

	// execute test cases
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			err := ws.WriteMessage(tt.giveType, []byte(tt.giveStr))
			require.NoError(t, err, "ws write should not cause an error")

			_, resp, err := ws.ReadMessage()
			require.NoError(t, err, "ws read should not cause an error")
			assert.JSONEq(t, tt.wantStr, string(resp), "response should match")
		})
	}
}

func TestServer_Auth(t *testing.T) {
	// define cases
	tests := []struct {
		desc     string
		giveAuth string
		wantErr  error
	}{
		{
			desc:     "no auth header",
			giveAuth: "",
			wantErr:  websocket.ErrBadHandshake,
		},
		{
			desc:     "invalid auth header",
			giveAuth: "Basic kDgmmNnabzatzZmvAV",
			wantErr:  websocket.ErrBadHandshake,
		},
		{
			desc:     "correct auth header",
			giveAuth: "Basic dGVzdHVzZXI6dGVzdA==",
			wantErr:  nil,
		},
	}

	url := "ws://127.0.0.1:40100/channel"

	// run tests
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			header := make(http.Header)
			if tt.giveAuth != "" {
				header.Add("Authorization", tt.giveAuth)
			}
			ws, _, err := websocket.DefaultDialer.Dial(url, header)
			require.Equal(t, tt.wantErr, err)

			if ws != nil {
				ws.Close()
			}
		})
	}
}
