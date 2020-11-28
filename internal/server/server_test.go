package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nsf/jsondiff"
	log "github.com/sirupsen/logrus"
	conf "github.com/spf13/viper"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestWebsocket(t *testing.T) {
	conf.Set("users.testuser",
		"09d9623a149a4a0c043befcb448c9c3324be973230188ba412c008a2929f31d0")

	// create test server
	s := New("127.0.0.1:8080")
	s.Bind("/channel")
	time.Sleep(10 * time.Millisecond)

	// connect websocket
	header := make(http.Header)
	header.Add("Authorization", "Basic dGVzdHVzZXI6dGVzdA==")
	url := "ws://127.0.0.1:8080/channel"
	ws, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// define cases
	cases := []struct {
		desc string
		data string
		exp  string
	}{
		{
			desc: "sdp echo",
			data: `{
				"type":"sdp", 
				"src":"testuser", 
				"dst":"testuser", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
			exp: `{
				"type":"sdp", 
				"src":"testuser", 
				"dst":"testuser", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
		},
	}

	diffOpts := jsondiff.DefaultConsoleOptions()

	// execute test cases
	for _, c := range cases {
		if err := ws.WriteMessage(websocket.TextMessage, []byte(c.data)); err != nil {
			t.Errorf("%v: %v", c.desc, err)
		}
		_, got, err := ws.ReadMessage()
		if err != nil {
			t.Errorf("%v: %v", c.desc, err)
		}

		res, diff := jsondiff.Compare(got, []byte(c.exp), &diffOpts)
		if res != jsondiff.FullMatch {
			t.Errorf("%v: diff: %v", c.desc, diff)
		}
	}

	// close websocket
	ws.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	s.Shutdown()
}
