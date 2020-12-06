package client

import (
	"github.com/lx7/devnet/proto"
	log "github.com/sirupsen/logrus"
)

// stateFunc represents a state and implements state transitions by returning
// another stateFunc for the following state.
type stateFunc func(*Controller) stateFunc

func stateStarting(c *Controller) stateFunc {
	log.Info("entering state: STARTING")

	select {
	case e := <-c.events:
		switch e {
		case evInitialized:
			return stateIdle
		default:
			return stateStarting
		}
	}
}

func stateIdle(c *Controller) stateFunc {
	log.Info("entering state: IDLE")

	select {
	case e := <-c.events:
		switch e {
		case evOfferSent:
			return stateCalling
		}
	case sig := <-c.signal.Receive():
		switch m := sig.(type) {
		case *proto.SDPMessage:
			session, err := NewSession(m.Src)
			if err != nil {
				log.Error("create session: ", err)
				return stateIdle
			}
			c.session = session
			if err := c.session.SetRemoteDescription(m.SDP); err != nil {
				log.Error("set remote description: ", err)
				return stateIdle
			}

			// accept connection
			answer, err := c.session.CreateAnswer()
			if err != nil {
				log.Error("create answer: ", err)
				return stateIdle
			}
			if err := c.signal.Send(&proto.SDPMessage{
				Src: c.user,
				Dst: c.session.Peer,
				SDP: answer,
			}); err != nil {
				log.Error("send answer: ", err)
				return stateIdle
			}
			c.session.StartOutboundPipes()
			return stateConnected
		}
	}
	return stateIdle
}

func stateCalling(c *Controller) stateFunc {
	log.Info("entering state: CALLING")

	select {
	case e := <-c.events:
		switch e {
		case evHangup:
			// TODO: clean up connection state
			c.session.Close()
			return stateIdle
		}

	case sig := <-c.signal.Receive():
		switch m := sig.(type) {
		case *proto.SDPMessage:
			if err := c.session.SetRemoteDescription(m.SDP); err != nil {
				log.Error("set remote description: ", err)
				return stateIdle
			}
			c.session.StartOutboundPipes()
			return stateConnected
		}
	}
	return stateCalling
}

func stateConnected(c *Controller) stateFunc {
	log.Info("entering state: CONNECTED")

	select {
	case e := <-c.events:
		switch e {
		case evHangup:
			c.session.Close()
			// TODO: clean up connection state
			return stateIdle
		}
	case <-c.signal.Receive():
	}
	return stateConnected
}

func stateClosed(c *Controller) stateFunc {
	return nil
}
