package client

import (
	"reflect"

	"github.com/lx7/devnet/proto"
	"github.com/pion/webrtc/v3"
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
	config  webrtc.Configuration
	forward chan *proto.Frame

	h       map[reflect.Type]handler
	sevents chan Event
	pevents chan Event
	done    chan bool
}

func NewSession(self string, signal SignalSendReceiver) (*DefaultSession, error) {
	s := DefaultSession{
		Self: self,

		signal:  signal,
		peers:   make(map[string]Peer),
		forward: make(chan *proto.Frame, 10),

		h:       make(map[reflect.Type]handler),
		pevents: make(chan Event, 10),
		sevents: make(chan Event, 10),
		done:    make(chan bool),
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

			switch pl := frame.Payload.(type) {
			case *proto.Frame_Config:
				log.Info().Stringer("config", pl.Config).Msg("config update")
				if pl.Config.Webrtc == nil {
					continue
				}
				if len(pl.Config.Webrtc.Iceservers) == 0 {
					continue
				}

				s.config = webrtc.Configuration{
					ICEServers: []webrtc.ICEServer{{
						URLs: []string{pl.Config.Webrtc.Iceservers[0].Url},
					}},
				}

			case *proto.Frame_Ice, *proto.Frame_Sdp:
				if frame.Dst != s.Self {
					log.Warn().Msg("received sdp message for self")
					continue
				}
				p, ok := s.peers[frame.Src]
				if !ok {
					var err error
					p, err = NewPeer(PeerOptions{
						Name:    frame.Src,
						Config:  s.config,
						Signals: s.forward,
						Events:  s.pevents,
					})
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
		case e := <-s.pevents:
			switch e := e.(type) {
			case EventPeerConnected, EventStreamStart, EventStreamEnd:
				s.sevents <- e
			case EventPeerDisconnected:
				delete(s.peers, e.Peer.Name())
				s.sevents <- e
			}
		case <-s.done:
			break
		}
	}
	log.Info().Msg("session closed")
}

func (s *DefaultSession) Connect(name string) error {
	if p, ok := s.peers[name]; ok && p != nil {
		p.Close()
	}

	p, err := NewPeer(PeerOptions{
		Name:    name,
		Config:  s.config,
		Signals: s.forward,
		Events:  s.pevents,
	})
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
	close(s.sevents)
	close(s.pevents)
	close(s.done)

	log.Info().Str("self", s.Self).Msg("session closed")
}

func (s *DefaultSession) Events() <-chan Event {
	return s.sevents
}

func (s *DefaultSession) handleSignalStateChange(st SignalState) {
	log.Info().Stringer("state", st).Msg("signaling: connection state changed")
	switch st {
	case SignalStateConnected:
		s.sevents <- EventConnected{}
	case SignalStateDisconnected:
		s.sevents <- EventDisconnected{}
	}
}
