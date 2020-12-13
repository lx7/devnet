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
	case frame := <-c.signal.Receive():
		switch p := frame.Payload.(type) {
		case *proto.Frame_Sdp:
			session, err := NewSession(frame.Src)
			if err != nil {
				log.Error("create session: ", err)
				return stateIdle
			}
			c.session = session
			offerSD := proto.ToPion(p)
			if err := c.session.SetRemoteDescription(offerSD); err != nil {
				log.Error("set remote description: ", err)
				return stateIdle
			}

			// accept connection
			answerSD, err := c.session.CreateAnswer()
			if err != nil {
				log.Error("create answer: ", err)
				return stateIdle
			}

			frame := &proto.Frame{
				Src:     c.user,
				Dst:     c.session.Peer,
				Payload: proto.WithPion(answerSD),
			}
			if err = c.signal.Send(frame); err != nil {
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

	case frame := <-c.signal.Receive():
		switch p := frame.Payload.(type) {
		case *proto.Frame_Sdp:
			responseSD := proto.ToPion(p)
			if err := c.session.SetRemoteDescription(responseSD); err != nil {
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
