package testutil

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{}

func Echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("testutil echohandler upgrade")
	}
	defer c.Close()

	for {
		mt, m, err := c.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Fatal().Err(err).Msg("testutil echohandler close error")
			}
			return
		}
		if err := c.WriteMessage(mt, m); err != nil {
			log.Fatal().Err(err).Msg("testutil echohandler write error")
		}
	}
}
