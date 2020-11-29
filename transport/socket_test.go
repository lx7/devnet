package transport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestSocket(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(echo))
	defer server.Close()

	url := "ws" + strings.TrimPrefix(server.URL, "http")

	socket := Dial(url, nil)
	time.Sleep(100 * time.Millisecond)
	defer socket.Close()

	msg := []byte("Lorem ipsum dolor sit amet {} -.,)(")

	for i := 0; i < 10; i++ {
		if err := socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			t.Fatalf("%v", err)
		}
		_, got, err := socket.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(got) != string(msg) {
			t.Fatalf("message mismatch: exp: '%s' got: '%s'", msg, got)
		}
	}
}

func echo(w http.ResponseWriter, r *http.Request) {
	socket, err := Upgrade(w, r)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer socket.Close()

	for {
		mt, message, err := socket.ReadMessage()
		if err != nil {
			return
		}

		if err := socket.WriteMessage(mt, message); err != nil {
			return
		}
	}
}
