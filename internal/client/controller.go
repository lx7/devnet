package client

import (
	"fmt"
	"sync"

	"github.com/lx7/devnet/proto"
	log "github.com/sirupsen/logrus"
)

type event int

const (
	evInitialized event = iota
	evOfferSent
	evHangup
)

// Controller represents the local client and controls processing of events
// and media streams.
type Controller struct {
	user         string
	pass         string
	signalingURL string

	signal  SignalSendReceiver
	session *Session

	events chan event
	done   chan bool
}

func NewController(s SignalSendReceiver, user string) (*Controller, error) {
	c := Controller{
		user:   user,
		signal: s,

		events: make(chan event),
		done:   make(chan bool, 2),
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
	c.session.StartGStreamer()
	wg.Wait()
}

// StartShare initiates screen sharing.
func (c *Controller) StartShare(dst string) error {
	log.Info("calling peer: ", dst)
	session, err := NewSession(dst)
	if err != nil {
		return fmt.Errorf("new session: %v", err)
	}
	c.session = session

	offer, err := c.session.CreateOffer()
	if err != nil {
		return fmt.Errorf("share offer: %v", err)
	}

	err = c.signal.Send(&proto.SDPMessage{
		Src: c.user,
		Dst: dst,
		SDP: offer,
	})
	if err != nil {
		return fmt.Errorf("send offer: %v", err)
	}
	c.events <- evOfferSent
	return nil
}
