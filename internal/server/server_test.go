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
	log.SetLevel(log.ErrorLevel)
}

func TestServer(t *testing.T) {
	conf.Set("users.testuser",
		"09d9623a149a4a0c043befcb448c9c3324be973230188ba412c008a2929f31d0")

	// create test server
	s := New("127.0.0.1:40100")
	s.Serve("/channel")
	time.Sleep(10 * time.Millisecond)

	// connect websocket
	header := make(http.Header)
	header.Add("Authorization", "Basic dGVzdHVzZXI6dGVzdA==")
	url := "ws://127.0.0.1:40100/channel"
	ws, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// define cases
	tests := []struct {
		desc string
		give string
		want string
	}{
		{
			desc: "sdp echo",
			give: `{
				"type":"sdp", 
				"src":"testuser", 
				"dst":"testuser", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
			want: `{
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
			err := ws.WriteMessage(websocket.TextMessage, []byte(tt.give))
			require.NoError(t, err, "ws write should not cause an error")

			_, resp, err := ws.ReadMessage()
			require.NoError(t, err, "ws read should not cause an error")
			assert.JSONEq(t, tt.want, string(resp), "response should match")
		})
	}

	// close websocket
	ws.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	s.Shutdown()
}
