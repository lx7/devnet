package testutil

import (
	"net/http"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{}

func Echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal("testutil echohandler upgrade error: ", err)
	}
	defer c.Close()

	for {
		mt, m, err := c.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Fatal("testutil echohandler unexpected close error: ", err)
			}
			return
		}
		if err := c.WriteMessage(mt, m); err != nil {
			log.Fatal("testutil echohandler write error: ", err)
		}
	}
}
