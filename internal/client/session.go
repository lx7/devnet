package client

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/gst"
	"github.com/lx7/devnet/proto"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
	conf "github.com/spf13/viper"
)

const (
	LocalScreen int = iota
	LocalCamera
	LocalVoice
	_maxLocalStreams
)

const (
	RemoteScreen int = iota
	RemoteCamera
	RemoteVoice
	_maxRemoteStreams
)

type SessionI interface {
	Connect(peer string) error
	SetOverlay(int, *gtk.DrawingArea)
	StartStream(int)
	Events() <-chan Event
}

type Session struct {
	Self, Peer string

	conn   *webrtc.PeerConnection
	signal SignalSendReceiver
	pdata  proto.FrameSendReceiver

	h      map[reflect.Type]handler
	events chan Event
	ls     []*LocalStream
	rs     []*RemoteStream
}

func NewSession(self string, signal SignalSendReceiver) (*Session, error) {
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, fmt.Errorf("new peer connection: %v", err)

	}
	s := Session{
		Self:   self,
		signal: signal,
		conn:   conn,
		ls:     make([]*LocalStream, _maxLocalStreams),
		rs:     make([]*RemoteStream, _maxRemoteStreams),
		events: make(chan Event, 5),
		h:      make(map[reflect.Type]handler),
	}

	s.signal.HandleStateChange(s.handleSignalStateChange)
	s.conn.OnICECandidate(s.handleICECandidate)
	s.conn.OnICEConnectionStateChange(s.handleICEStateChange)
	s.conn.OnSignalingStateChange(s.handlePionSignalingStateChange)
	s.conn.OnTrack(s.handleTrack)

	voicePreset, err := gst.GetPreset(
		gst.Voice,
		gst.Opus,
		gst.NewHardwareCodec(""),
	)
	if err != nil {
		return nil, err
	}
	screenPreset, err := gst.GetPreset(
		gst.Screen,
		gst.H264,
		gst.NewHardwareCodec(conf.GetString("video.hardware")),
	)
	if err != nil {
		return nil, err
	}

	s.ls[LocalVoice], err = NewLocalStream(s.conn, &LocalStreamOpts{
		ID:     "audio",
		Group:  "chat",
		Preset: voicePreset,
	})
	if err != nil {
		return nil, err
	}
	s.ls[LocalScreen], err = NewLocalStream(s.conn, &LocalStreamOpts{
		ID:     "video",
		Group:  "screen",
		Preset: screenPreset,
	})
	if err != nil {
		return nil, err
	}

	s.rs[RemoteVoice], err = NewRemoteStream(s.conn, RemoteStreamOpts{
		ID:     "audio",
		Group:  "chat",
		Preset: voicePreset,
	})
	if err != nil {
		return nil, err
	}
	s.rs[RemoteScreen], err = NewRemoteStream(s.conn, RemoteStreamOpts{
		ID:     "video",
		Group:  "screen",
		Preset: screenPreset,
	})
	if err != nil {
		return nil, err
	}

	s.pdata, err = NewDataChannel(s.conn)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Session) Connect(peer string) error {
	if peer == s.Self {
		return fmt.Errorf("peer name equals self")
	}
	if s.conn == nil {
		return fmt.Errorf("webrtc connection not initialized")
	}

	offer, err := s.conn.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("create offer: %v", err)
	}

	err = s.conn.SetLocalDescription(offer)
	if err != nil {
		return fmt.Errorf("set local description to offer: %v", err)
	}

	frame := &proto.Frame{
		Src:     s.Self,
		Dst:     peer,
		Payload: proto.PayloadWithSD(offer),
	}
	if err = s.signal.Send(frame); err != nil {
		return fmt.Errorf("send frame: %v", err)
	}

	s.Peer = peer
	return nil
}

func (s *Session) Run() {
	for {
		select {
		case frame := <-s.signal.Receive():
			log.Trace().
				Str("src", frame.Src).
				Str("dst", frame.Dst).
				Stringer("type", reflect.TypeOf(frame.Payload)).
				Msg("sinaling frame received")

			switch p := frame.Payload.(type) {
			case *proto.Frame_Config:
				err := s.conn.SetConfiguration(webrtc.Configuration{
					ICEServers: []webrtc.ICEServer{{
						URLs: []string{p.Config.Webrtc.Iceservers[0].Url},
					}},
				})
				if err != nil {
					log.Error().Err(err).Msg("configure webrtc")
					continue
				}

			case *proto.Frame_Ice:
				if frame.Dst != s.Self {
					log.Warn().Msg("received ice candidate for self")
					continue
				}
				s.conn.AddICECandidate(p.ICECandidate())

			case *proto.Frame_Sdp:
				if frame.Dst != s.Self {
					log.Warn().Msg("received sdp message for self")
					continue
				}

				switch p.Sdp.Type {
				case proto.SDP_OFFER:
					s.Peer = frame.Src
					err := s.conn.SetRemoteDescription(p.SessionDescription())
					if err != nil {
						log.Error().Err(err).Msg("set remote session to offer")
						continue
					}

					answer, err := s.conn.CreateAnswer(nil)
					if err != nil {
						log.Error().Err(err).Msg("create answer")
						continue
					}

					err = s.conn.SetLocalDescription(answer)
					if err != nil {
						log.Error().
							Err(err).
							Str("self", s.Self).
							Msg("set local session description")
						continue
					}

					frame := &proto.Frame{
						Src:     s.Self,
						Dst:     s.Peer,
						Payload: proto.PayloadWithSD(answer),
					}
					if err = s.signal.Send(frame); err != nil {
						log.Error().Err(err).Msg("send answer")
						continue
					}
				case proto.SDP_ANSWER:
					err := s.conn.SetRemoteDescription(p.SessionDescription())
					if err != nil {
						log.Error().
							Err(err).
							Str("self", s.Self).
							Msg("set remote session to answer")
						continue
					}
				}
			}
		case frame := <-s.pdata.Receive():
			log.Trace().
				Stringer("type", reflect.TypeOf(frame.Payload)).
				Msg("frame received on data channel")
			switch p := frame.Payload.(type) {
			case *proto.Frame_Control:
				s.events <- EventRCon{
					Data: p.Control,
				}
			default:
				log.Error().
					Str("self", s.Self).
					Interface("payload", p).
					Msg("unknown payload type on data channel")

			}
		}
	}
	log.Info().Msg("session closed")

}

func (s *Session) StartStream(id int) {
	s.ls[id].Send()
}

func (s *Session) SetOverlay(id int, o *gtk.DrawingArea) {
	s.rs[id].SetOverlay(o)
}

func (s *Session) RCon(cm *proto.Control) error {
	log.Trace().Stringer("msg", cm).Msg("send control message to peer")

	frame := &proto.Frame{
		Payload: &proto.Frame_Control{cm},
	}

	err := s.pdata.Send(frame)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) handleSignalStateChange(st SignalState) {
	log.Info().Stringer("state", st).Msg("signal connection state changed")
	switch st {
	case SignalStateConnected:
		s.events <- EventConnected{}
	case SignalStateDisconnected:
		s.events <- EventDisconnected{}
	}
}

func (s *Session) handleICECandidate(c *webrtc.ICECandidate) {
	if c == nil {
		return
	}
	frame := &proto.Frame{
		Src:     s.Self,
		Dst:     s.Peer,
		Payload: proto.PayloadWithICECandidate(c.ToJSON()),
	}
	if err := s.signal.Send(frame); err != nil {
		log.Error().Err(err).Msg("send ICE candidate")
	}
}

func (s *Session) handlePionSignalingStateChange(st webrtc.SignalingState) {
	log.Trace().
		Str("self", s.Self).
		Stringer("state", st).
		Msg("webrtc signaling state changed")
}

func (s *Session) handleICEStateChange(st webrtc.ICEConnectionState) {
	log.Trace().
		Str("self", s.Self).
		Stringer("state", st).
		Msg("ICE connection state changed")
	switch st {
	case webrtc.ICEConnectionStateConnected:
		s.events <- EventSessionStart{}
		return
	case webrtc.ICEConnectionStateDisconnected:
		s.Close()
	case webrtc.ICEConnectionStateFailed:
	case webrtc.ICEConnectionStateClosed:
		s.events <- EventSessionEnd{}
	}
}

func (s *Session) handleTrack(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	log.Info().
		Str("self", s.Self).
		Str("track_id", track.ID()).
		Msg("new track")

	// temporary workaround until pion webrtc implements incoming RTCP events
	go func() {
		ticker := time.NewTicker(time.Second * 3)
		for range ticker.C {
			err := s.conn.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())},
			})
			if err == io.ErrClosedPipe {
				return
			} else if err != nil {
				log.Error().
					Str("track_id", track.ID()).
					Err(err).
					Msg("rctp send")
			}
		}
	}()

	/*
		codec, err := gst.CodecByName(t.Codec().Name)
		if err != nil {
			return nil, fmt.Errorf("codec for new track: %v", err)
		}
	*/
	switch track.ID() {
	case "audio":
		log.Trace().Str("track_id", track.ID()).Msg("starting receive")
		s.rs[RemoteVoice].Receive(track)
	case "video":
		log.Trace().Str("track_id", track.ID()).Msg("starting receive")
		s.events <- EventSCInboundStart{}
		s.rs[RemoteScreen].Receive(track)
	default:
		log.Error().Str("track_id", track.ID()).Msg("track id unknown")
		return
	}

}

func (s *Session) Close() {
	for _, stream := range s.rs {
		if stream == nil {
			continue
		}
		stream.Close()
	}
	for _, stream := range s.ls {
		if stream == nil {
			continue
		}
		stream.Close()
	}
	s.conn.Close()
}

func (s *Session) Events() <-chan Event {
	return s.events
}
