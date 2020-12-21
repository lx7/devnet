package client

import (
	"fmt"
	"sync"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/proto"
	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
)

type Event int

const (
	evUndefined Event = iota
	evInitialized
	evOfferSent
	evHangup
)

const (
	EventSCInboundStart Event = iota + 10
	EventSCInboundEnd
)

type HandlerFunc func()

type ScreenCaster interface {
	SetOverlayHandle(*gtk.DrawingArea)
	AddPeer(peer string) error
	Handle(Event, HandlerFunc)
	StartScreenCast()
	StopScreenCast()
}

// Controller represents the local client and controls processing of events
// and media streams.
type Controller struct {
	name   string
	signal SignalSendReceiver

	session   *Session
	wconf     webrtc.Configuration
	scOverlay *gtk.DrawingArea

	events chan Event
	done   chan bool
	h      map[Event]HandlerFunc
}

func NewController(s SignalSendReceiver, name string, conf webrtc.Configuration) (*Controller, error) {
	c := Controller{
		name:   name,
		signal: s,
		wconf:  conf,

		events: make(chan Event),
		done:   make(chan bool, 2),
		h:      make(map[Event]HandlerFunc),
	}
	return &c, nil
}

// Run starts processing of events and media streams. Must be locked to the
// main thread.
func (c *Controller) Run() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for state := stateStarting; state != nil; {
			state = state(c)
		}
	}()

	c.events <- evInitialized
	wg.Wait()
}

// AddPeer initiates a session with peer.
func (c *Controller) AddPeer(peer string) error {
	log.Info("calling peer: ", peer)
	session, err := NewSession(SessionOpts{
		Peer:              peer,
		wconf:             c.wconf,
		ScreenCastOverlay: c.scOverlay,
	})
	if err != nil {
		return fmt.Errorf("new session: %v", err)
	}
	c.session = session

	offer, err := c.session.CreateOffer()
	if err != nil {
		return fmt.Errorf("share offer: %v", err)
	}

	frame := &proto.Frame{
		Src:     c.name,
		Dst:     peer,
		Payload: &proto.Frame_Sdp{offer},
	}
	if err = c.signal.Send(frame); err != nil {
		return fmt.Errorf("send frame: %v", err)
	}
	c.events <- evOfferSent
	return nil
}

func (c *Controller) SetOverlayHandle(h *gtk.DrawingArea) {
	c.scOverlay = h
}

func (c *Controller) StartScreenCast() {
	c.session.ScreenCast.Start()
}

func (c *Controller) StopScreenCast() {
}

func (c *Controller) Handle(e Event, f HandlerFunc) {
	c.h[e] = f
}

func (c *Controller) notify(e Event) {
	f, ok := c.h[e]
	if ok {
		f()
	}
}
