package server

import (
	"github.com/lx7/devnet/proto"

	log "github.com/sirupsen/logrus"
)

// Switch provides message forwarding between clients.
type Switch interface {
	Register(Client)
	Unregister(Client)
	Forward() chan<- *proto.SDPMessage
	Run()
	Shutdown()
}

// DefaultSwitch implements the Switch interface.
type DefaultSwitch struct {
	clients map[string]Client

	forward    chan *proto.SDPMessage
	broadcast  chan *proto.SDPMessage
	register   chan Client
	unregister chan Client
	done       chan bool
}

// NewSwitch returns a new Switch instance.
func NewSwitch() *DefaultSwitch {
	return &DefaultSwitch{
		broadcast:  make(chan *proto.SDPMessage),
		forward:    make(chan *proto.SDPMessage),
		register:   make(chan Client),
		unregister: make(chan Client),
		clients:    make(map[string]Client),
		done:       make(chan bool),
	}
}

// Register connects c to the switch and starts message processing.
// Returns on termination of the client run loop.
func (sw *DefaultSwitch) Register(c Client) {
	sw.register <- c
	c.Attach(sw)
}

func (sw *DefaultSwitch) Unregister(c Client) {
	sw.unregister <- c
}

// Forward returns the switches forward channel.
func (sw *DefaultSwitch) Forward() chan<- *proto.SDPMessage {
	return sw.forward
}

// Run implements the message handling loop.
func (sw *DefaultSwitch) Run() {
	for {
		select {
		case client := <-sw.register:
			log.Info("registering new client: ", client.Name())
			sw.clients[client.Name()] = client
		case client := <-sw.unregister:
			if _, ok := sw.clients[client.Name()]; ok {
				log.Info("unregistering client: ", client.Name())
				delete(sw.clients, client.Name())
				close(client.Send())
			}
		case m := <-sw.broadcast:
			for _, client := range sw.clients {
				select {
				case client.Send() <- m:
				default:
					sw.unregister <- client
				}
			}
		case m := <-sw.forward:
			if client, ok := sw.clients[m.Dst]; ok {
				// TODO: verify sender
				log.Tracef("forwarding message: %s -> %s", m.Src, m.Dst)
				select {
				case client.Send() <- m:
				default:
					sw.unregister <- client
				}
			} else {
				log.Tracef("client %s absent, discarding message", m.Dst)
			}
		case <-sw.done:
			return
		}
	}
}

// Shutdown unregisters all clients and stops the run loop.
func (sw *DefaultSwitch) Shutdown() {
	for _, c := range sw.clients {
		sw.unregister <- c
	}
	close(sw.done)
}
