package signaling

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lx7/devnet/transport"

	"github.com/gorilla/websocket"
	"github.com/nsf/jsondiff"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

var sw *Switch

func TestClientAndSwitch(t *testing.T) {
	// start a test server
	server := httptest.NewServer(http.HandlerFunc(serve))
	url := "ws" + strings.TrimPrefix(server.URL, "http")

	// create a switch
	sw = NewSwitch()
	go sw.Run()

	// allow some time for startup
	time.Sleep(100 * time.Millisecond)

	// create sender and receiver connections
	sender := mkConn(t, url, "sender")
	receiver := mkConn(t, url, "receiver")

	// allow some time for connections being established
	time.Sleep(100 * time.Millisecond)

	// define cases
	cases := []struct {
		desc string
		data string
		exp  string
	}{
		{
			desc: "sdp forward",
			data: `{
				"type":"sdp", 
				"src":"sender", 
				"dst":"receiver", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
			exp: `{
				"type":"sdp", 
				"src":"sender", 
				"dst":"receiver", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
		},
	}

	diffOpts := jsondiff.DefaultConsoleOptions()

	// execute test cases
	for _, c := range cases {
		if err := sender.WriteMessage(websocket.TextMessage, []byte(c.data)); err != nil {
			t.Errorf("%v: %v", c.desc, err)
		}
		_, got, err := receiver.ReadMessage()
		if err != nil {
			t.Errorf("%v: %v", c.desc, err)
		}

		res, diff := jsondiff.Compare(got, []byte(c.exp), &diffOpts)
		if res != jsondiff.FullMatch {
			t.Errorf("%v: diff: %v", c.desc, diff)
		}
	}

	closeConn(t, sender)
	closeConn(t, receiver)

	// allow some time for connections to be terminated
	time.Sleep(100 * time.Millisecond)

	server.Close()
	sw.Shutdown()
}

func mkConn(t *testing.T, url, name string) *websocket.Conn {
	header := make(http.Header)
	header.Add("Name", name)

	s, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return s
}

func closeConn(t *testing.T, s *websocket.Conn) {
	s.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
}

func serve(w http.ResponseWriter, r *http.Request) {
	name := r.Header.Get("Name")

	socket, err := transport.Upgrade(w, r)
	if err != nil {
		log.Fatalf("%v", err)
	}

	c := NewClient(socket, sw, name)
	sw.Register(c)
	defer sw.Unregister(c)

	go c.WritePump()
	c.ReadPump()
}
