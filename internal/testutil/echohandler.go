package testutil

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func Echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	for {
		mt, m, err := c.ReadMessage()
		if err != nil {
			log.Fatal(err)
		}
		if err := c.WriteMessage(mt, m); err != nil {
			log.Fatal(err)
		}
	}
}
