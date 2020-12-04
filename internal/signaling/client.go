package signaling

import (
	"github.com/lx7/devnet/proto"
	"github.com/lx7/devnet/transport"

	log "github.com/sirupsen/logrus"
)

// Client implements the client connection handler per socket and attaches
// to a switch
type Client struct {
	name string
	sw   *Switch
	rw   transport.MessageReadWriter
	mux  *transport.Mux

	send chan *proto.SDPMessage
	done chan bool
}

// NewClient returns a new Client instance, attached to socket and sw
func NewClient(rw transport.MessageReadWriter, sw *Switch, name string) *Client {
	c := &Client{
		name: name,
		sw:   sw,
		rw:   rw,

		send: make(chan *proto.SDPMessage, 64),
		done: make(chan bool),
	}
	c.mux = &transport.Mux{
		ReadWriter: rw,
		Consumer:   c,
	}
	return c
}

// HandleSDP implements the transport.Consumer interface
func (c *Client) HandleSDP(m *proto.SDPMessage) {
	c.sw.forward <- m
}

// HandleClose implements the transport.Consumer interface
func (c *Client) HandleClose() {
	c.sw.unregister <- c
}

// Run implements the main loop for the client.
func (c *Client) Run() {
	defer c.rw.Close()
	go c.mux.Receive()
	c.writePump()
}

func (c *Client) writePump() {
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
		case <-c.done:
			return
		}
	}
}
