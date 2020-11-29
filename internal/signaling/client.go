package signaling

import (
	"github.com/lx7/devnet/proto"
	"github.com/lx7/devnet/transport"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Client implements the client connection handler per socket and attaches
// to a switch
type Client struct {
	name string
	sw   *Switch
	mux  *transport.Mux
	send chan *proto.SDPMessage
}

// NewClient returns a new Client instance, attached to socket and sw
func NewClient(socket *transport.Socket, sw *Switch, name string) *Client {
	c := &Client{
		name: name,
		sw:   sw,
		send: make(chan *proto.SDPMessage, 64),
	}
	c.mux = transport.NewMux(socket, c)
	return c
}

// ReadPump uses mux to read JSON messages from the socket and dispatches
// them as native entities to the transport.Consumer interface methods.
func (c *Client) ReadPump() {
	defer func() {
		c.sw.unregister <- c
		c.mux.Close()
	}()
	if err := c.mux.Receive(); err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Info("connection closed")
		} else {
			log.Errorf("readpump read from client '%s': %s", c.name, err)
		}
	}
}

// WritePump comsumes native messages from the send channel, serializes and
// forwards them through the websocket connection to the client.
func (c *Client) WritePump() {
	for {
		select {
		case m, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.mux.Send(m); err != nil {
				log.Errorf("writepump send to client '%s': %s", c.name, err)
				return
			}
		}
	}
}

// HandleSDP implements the transport.Consumer interface
func (c *Client) HandleSDP(m *proto.SDPMessage) {
	c.sw.forward <- m
}
