package signaling

import (
	"github.com/lx7/devnet/proto"
	"github.com/lx7/devnet/transport"

	log "github.com/sirupsen/logrus"
)

func NewSwitch() *Switch {
	return &Switch{
		broadcast:  make(chan *proto.SDPMessage),
		forward:    make(chan *proto.SDPMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		done:       make(chan bool),
	}
}

// Switch provides message forwarding between clients.
type Switch struct {
	clients map[string]*Client

	forward    chan *proto.SDPMessage
	broadcast  chan *proto.SDPMessage
	register   chan *Client
	unregister chan *Client
	done       chan bool
}

// Attach connects a MessageReadWriter to the switch and returns when mrw
// closes.
func (sw *Switch) Attach(rw transport.MessageReadWriter, name string) {
	c := NewClient(rw, sw, name)
	sw.register <- c
	c.Run()
}

// Run implements the message handling loop.
func (sw *Switch) Run() {
	for {
		select {
		case client := <-sw.register:
			log.Info("registering new client: ", client.name)
			sw.clients[client.name] = client
		case client := <-sw.unregister:
			if _, ok := sw.clients[client.name]; ok {
				log.Info("unregistering client: ", client.name)
				delete(sw.clients, client.name)
				close(client.send)
			}
		case message := <-sw.broadcast:
			for _, client := range sw.clients {
				select {
				case client.send <- message:
				default:
					sw.unregister <- client
				}
			}
		case message := <-sw.forward:
			if client, ok := sw.clients[message.Dst]; ok {
				// TODO: verify sender
				log.Tracef("forwarding message: %s -> %s",
					message.Src, message.Dst)
				select {
				case client.send <- message:
				default:
					sw.unregister <- client
				}
			} else {
				log.Tracef("client absent, discarding message: %s -> %s",
					message.Src, message.Dst)
			}
		case <-sw.done:
			return
		}
	}
}

func (sw *Switch) Shutdown() {
	for _, c := range sw.clients {
		sw.unregister <- c
	}
}
