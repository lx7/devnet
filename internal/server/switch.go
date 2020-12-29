package server

import (
	"github.com/lx7/devnet/proto"

	"github.com/rs/zerolog/log"
)

// Switch provides message forwarding between clients.
type Switch interface {
	Register(Client)
	Unregister(Client)
	Forward() chan<- *proto.Frame
	Run()
	Shutdown()
}

// DefaultSwitch implements the Switch interface.
type DefaultSwitch struct {
	clients map[string]Client

	forward    chan *proto.Frame
	broadcast  chan *proto.Frame
	register   chan Client
	unregister chan Client
	done       chan bool
}

// NewSwitch returns a new Switch instance.
func NewSwitch() *DefaultSwitch {
	return &DefaultSwitch{
		broadcast:  make(chan *proto.Frame),
		forward:    make(chan *proto.Frame),
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
func (sw *DefaultSwitch) Forward() chan<- *proto.Frame {
	return sw.forward
}

// Run implements the message handling loop.
func (sw *DefaultSwitch) Run() {
	for {
		select {
		case client := <-sw.register:
			log.Info().Str("user", client.Name()).Msg("registering client")
			sw.clients[client.Name()] = client
		case client := <-sw.unregister:
			if _, ok := sw.clients[client.Name()]; ok {
				log.Info().Str("user", client.Name()).Msg("unregistering client")
				delete(sw.clients, client.Name())
				close(client.Send())
			}
		case f := <-sw.broadcast:
			for _, client := range sw.clients {
				select {
				case client.Send() <- f:
				default:
					sw.unregister <- client
				}
			}
		case f := <-sw.forward:
			if client, ok := sw.clients[f.Dst]; ok {
				// TODO: verify sender
				log.Trace().
					Str("src", f.Src).
					Str("dst", f.Dst).
					Msg("forwarding message")
				select {
				case client.Send() <- f:
				default:
					sw.unregister <- client
				}
			} else {
				log.Trace().
					Str("src", f.Src).
					Str("dst", f.Dst).
					Msg("client absent, discarding message")
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
