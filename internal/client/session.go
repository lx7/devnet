package client

import (
	"errors"
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
	RemoteScreen
	RemoteCamera
	RemoteVoice
	_maxStreams
)

type SessionI interface {
	Connect(peer string) error
	SetOverlay(int, *gtk.GLArea)
	StartStream(int) error
	StopStream(int) error
	Events() <-chan Event
}

type Session struct {
	Self, Peer string

	conn   *webrtc.PeerConnection
	signal SignalSendReceiver
	pdata  proto.FrameSendReceiver

	h       map[reflect.Type]handler
	events  chan Event
	streams []Stream
}

func NewSession(self string, signal SignalSendReceiver) (*Session, error) {
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, fmt.Errorf("new peer connection: %v", err)

	}
	s := Session{
		Self:    self,
		signal:  signal,
		conn:    conn,
		h:       make(map[reflect.Type]handler),
		events:  make(chan Event, 5),
		streams: make([]Stream, _maxStreams),
	}

	s.signal.HandleStateChange(s.handleSignalStateChange)
	s.conn.OnICECandidate(s.handleICECandidate)
	s.conn.OnICEConnectionStateChange(s.handleICEStateChange)
	s.conn.OnSignalingStateChange(s.handlePionSignalingStateChange)
	s.conn.OnTrack(s.handleTrack)

	if err := s.initStreams(); err != nil {
		return nil, err
	}

	s.pdata, err = NewDataChannel(s.conn)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (s *Session) initStreams() error {
	for _, st := range []struct {
		source   string
		mimetype string
		lsid     int
		rsid     int
	}{
		{
			source:   gst.Voice,
			mimetype: gst.MimeTypeOpus,
			lsid:     LocalVoice,
			rsid:     RemoteVoice,
		},
		{
			source:   gst.Camera,
			mimetype: gst.MimeTypeH264,
			lsid:     LocalCamera,
			rsid:     RemoteCamera,
		},
		{
			source:   gst.Screen,
			mimetype: gst.MimeTypeH264,
			lsid:     LocalScreen,
			rsid:     RemoteScreen,
		},
	} {
		codec, err := gst.GetPreset(
			st.source,
			st.mimetype,
			gst.NewHardwareCodec(conf.GetString(st.source+".hardware")),
		)
		if err != nil {
			return err
		}

		if p := conf.GetString("codec." + st.source + ".local"); p != "" {
			codec.Local = p
		}
		s.streams[st.lsid], err = NewStreamLocal(s.conn, &StreamOpts{
			ID:       st.source,
			Group:    "devnet",
			MimeType: st.mimetype,
			Pipeline: codec.Local,
		})
		if err != nil {
			return err
		}

		if p := conf.GetString("codec." + st.source + ".remote"); p != "" {
			codec.Remote = p
		}
		s.streams[st.rsid], err = NewStreamRemote(s.conn, StreamOpts{
			ID:       st.source,
			Group:    "devnet",
			Pipeline: codec.Remote,
		})
		if err != nil {
			return err
		}
	}

	return nil
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
				if p.Config.Webrtc == nil {
					continue
				}

				if len(p.Config.Webrtc.Iceservers) == 0 {
					continue
				}

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

func (s *Session) StartStream(id int) error {
	sender, ok := s.streams[id].(StreamSender)
	if !ok {
		return errors.New("stream is not local")
	}

	log.Info().Int("id", id).Msg("start stream")
	sender.Send()

	return nil
}

func (s *Session) StopStream(id int) error {
	sender, ok := s.streams[id].(StreamSender)
	if !ok {
		return errors.New("stream is not local")
	}

	log.Info().Int("id", id).Msg("stop stream")
	sender.Stop()

	return nil
}

func (s *Session) SetOverlay(id int, o *gtk.GLArea) {
	s.streams[id].SetOverlay(o)
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
	log.Info().Stringer("state", st).Msg("signaling: connection state changed")
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
		log.Info().Str("self", s.Self).Msg("peer connection established")
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

	var st Stream
	switch track.ID() {
	case gst.Voice:
		st = s.streams[RemoteVoice]
	case gst.Camera:
		s.events <- EventCameraInboundStart{}
		st = s.streams[RemoteCamera]
	case gst.Screen:
		s.events <- EventSCInboundStart{}
		st = s.streams[RemoteScreen]
	default:
		log.Error().Str("track_id", track.ID()).Msg("track id unknown")
		return
	}

	stream, ok := st.(StreamReceiver)
	if !ok {
		log.Error().Str("track_id", track.ID()).Msg("stream is not remote")
		return
	}

	log.Trace().Str("track_id", track.ID()).Msg("starting receive")
	stream.Receive(track)
}

func (s *Session) Close() {
	for _, stream := range s.streams {
		if stream == nil {
			continue
		}
		stream.Close()
	}
	s.pdata.Close()
	s.conn.Close()
	log.Info().Str("self", s.Self).Msg("peer connection closed")
}

func (s *Session) Events() <-chan Event {
	return s.events
}
