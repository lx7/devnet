package client

import (
	"reflect"

	"github.com/lx7/devnet/proto"
	"github.com/rs/zerolog/log"
)

const (
	LocalScreen int = iota
	LocalCamera
	LocalVoice
	RemoteScreen
	RemoteCamera
	RemoteVoice
	_maxStreams
)

type Session interface {
	Connect(peer string) error
	Events() <-chan Event
}

type DefaultSession struct {
	Self string

	signal  SignalSendReceiver
	peers   map[string]Peer
	forward chan *proto.Frame

	h      map[reflect.Type]handler
	events chan Event
	done   chan bool
}

func NewSession(self string, signal SignalSendReceiver) (*DefaultSession, error) {
	s := DefaultSession{
		Self: self,

		signal:  signal,
		peers:   make(map[string]Peer),
		forward: make(chan *proto.Frame, 10),

		h:      make(map[reflect.Type]handler),
		events: make(chan Event, 10),
		done:   make(chan bool),
	}

	s.signal.HandleStateChange(s.handleSignalStateChange)

	return &s, nil
}

func (s *DefaultSession) Run() {
	for {
		select {
		case frame := <-s.signal.Receive():
			log.Trace().
				Str("src", frame.Src).
				Str("dst", frame.Dst).
				Stringer("type", reflect.TypeOf(frame.Payload)).
				Msg("sinaling frame received")

			switch frame.Payload.(type) {
			case *proto.Frame_Config:
				for name, peer := range s.peers {
					err := peer.HandleSignaling(frame)
					if err != nil {
						log.Error().Err(err).Str("peer", name).Msg("configure webrtc")
					}
				}

			case *proto.Frame_Ice, *proto.Frame_Sdp:
				if frame.Dst != s.Self {
					log.Warn().Msg("received sdp message for self")
					continue
				}
				p, ok := s.peers[frame.Src]
				if !ok {
					var err error
					p, err = NewPeer(frame.Src, s.forward, s.events)
					if err != nil {
						log.Error().Err(err).Str("peer", frame.Src).Msg("new peer")
						continue
					}
				}
				err := p.HandleSignaling(frame)
				if err != nil {
					log.Error().Err(err).Str("peer", frame.Src).Msg("handle frame")
					continue
				}
				s.peers[frame.Src] = p
			}
		case frame := <-s.forward:
			frame.Src = s.Self
			if err := s.signal.Send(frame); err != nil {
				log.Error().Err(err).Str("dst", frame.Dst).Msg("send frame")
				continue
			}
		// TODO: handle peer events
		// TODO: remove peer on close
		case <-s.done:
			break
		}
	}
	log.Info().Msg("session closed")
}

func (s *DefaultSession) Connect(name string) error {
	if p, ok := s.peers[name]; ok {
		p.Close()
	}

	p, err := NewPeer(name, s.forward, s.events)
	if err != nil {
		return err
	}
	p.Connect()
	s.peers[name] = p

	return nil
}

func (s *DefaultSession) Close() {
	for _, peer := range s.peers {
		if peer == nil {
			continue
		}
		peer.Close()
		peer = nil
	}
	close(s.events)
	close(s.done)

	log.Info().Str("self", s.Self).Msg("session closed")
}

func (s *DefaultSession) Events() <-chan Event {
	return s.events
}

func (s *DefaultSession) handleSignalStateChange(st SignalState) {
	log.Info().Stringer("state", st).Msg("signaling: connection state changed")
	switch st {
	case SignalStateConnected:
		s.events <- EventConnected{}
	case SignalStateDisconnected:
		s.events <- EventDisconnected{}
	}
}
