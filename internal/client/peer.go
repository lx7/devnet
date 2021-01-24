package client

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/lx7/devnet/gst"
	"github.com/lx7/devnet/proto"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
	conf "github.com/spf13/viper"
)

type Peer interface {
	VideoLocal() StreamSender
	VideoRemote() StreamReceiver
	AudioLocal() StreamSender
	AudioRemote() StreamReceiver
	ScreenLocal() StreamSender
	ScreenRemote() StreamReceiver

	Name() string
	HandleSignaling(*proto.Frame) error
	Close()
}

type DefaultPeer struct {
	name string
	conn *webrtc.PeerConnection

	signals chan<- *proto.Frame
	events  chan<- Event
	str     map[string]StreamReceiver
	stl     map[string]StreamSender
	data    proto.FrameSendReceiver
	done    chan bool
}

func NewPeer(name string, sig chan<- *proto.Frame, ev chan<- Event) (*DefaultPeer, error) {
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, fmt.Errorf("new peer connection: %v", err)

	}
	p := DefaultPeer{
		conn:    conn,
		name:    name,
		signals: sig,
		events:  ev,
		str:     make(map[string]StreamReceiver),
		stl:     make(map[string]StreamSender),
		done:    make(chan bool),
	}

	p.conn.OnICECandidate(p.handleICECandidate)
	p.conn.OnICEConnectionStateChange(p.handleICEStateChange)
	p.conn.OnTrack(p.handleTrack)

	if err := p.initStreams(); err != nil {
		return nil, err
	}

	p.data, err = NewDataChannel(p.conn)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *DefaultPeer) initStreams() error {
	for _, st := range []struct {
		source   string
		mimetype string
		id       string
	}{
		{
			source:   gst.Voice,
			mimetype: gst.MimeTypeOpus,
			id:       "audio",
		},
		{
			source:   gst.Camera,
			mimetype: gst.MimeTypeH264,
			id:       "video",
		},
		{
			source:   gst.Screen,
			mimetype: gst.MimeTypeH264,
			id:       "screen",
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

		if pd := conf.GetString("codec." + st.source + ".local"); pd != "" {
			codec.Local = pd
		}
		p.stl[st.id], err = NewStreamLocal(p.conn, &StreamOpts{
			ID:       st.id,
			Group:    "devnet",
			MimeType: st.mimetype,
			Pipeline: codec.Local,
		})
		if err != nil {
			return err
		}

		if pd := conf.GetString("codec." + st.source + ".remote"); pd != "" {
			codec.Remote = pd
		}
		p.str[st.id], err = NewStreamRemote(p.conn, StreamOpts{
			ID:       st.id,
			Group:    "devnet",
			Pipeline: codec.Remote,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *DefaultPeer) Connect() error {
	offer, err := p.conn.CreateOffer(nil)
	if err != nil {
		return fmt.Errorf("create offer: %v", err)
	}

	err = p.conn.SetLocalDescription(offer)
	if err != nil {
		return fmt.Errorf("set local description to offer: %v", err)
	}

	frame := &proto.Frame{
		Dst:     p.name,
		Payload: proto.PayloadWithSD(offer),
	}

	p.signals <- frame
	return nil
}

func (p *DefaultPeer) HandleSignaling(frame *proto.Frame) error {
	switch pl := frame.Payload.(type) {
	case *proto.Frame_Config:
		if pl.Config.Webrtc == nil {
			return nil
		}
		if len(pl.Config.Webrtc.Iceservers) == 0 {
			return nil
		}

		err := p.conn.SetConfiguration(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{{
				URLs: []string{pl.Config.Webrtc.Iceservers[0].Url},
			}},
		})
		if err != nil {
			return fmt.Errorf("configure webrtc: %v", err)
		}

	case *proto.Frame_Ice:
		p.conn.AddICECandidate(pl.ICECandidate())

	case *proto.Frame_Sdp:
		switch pl.Sdp.Type {
		case proto.SDP_OFFER:
			err := p.conn.SetRemoteDescription(pl.SessionDescription())
			if err != nil {
				return fmt.Errorf("set remote session to offer: %v", err)
			}

			answer, err := p.conn.CreateAnswer(nil)
			if err != nil {
				return fmt.Errorf("create answer: %v", err)
			}

			err = p.conn.SetLocalDescription(answer)
			if err != nil {
				return fmt.Errorf("set local session description: %v", err)
			}

			p.signals <- &proto.Frame{
				Dst:     p.name,
				Payload: proto.PayloadWithSD(answer),
			}
		case proto.SDP_ANSWER:
			err := p.conn.SetRemoteDescription(pl.SessionDescription())
			if err != nil {
				return fmt.Errorf("set remote session to answer: %v", err)
			}
		}
	}

	return nil
}

func (p *DefaultPeer) Run() {
	for {
		select {
		case frame := <-p.data.Receive():
			log.Trace().
				Stringer("type", reflect.TypeOf(frame.Payload)).
				Msg("frame received on data channel")
			switch pl := frame.Payload.(type) {
			case *proto.Frame_Control:
				p.events <- EventRCon{
					Peer: p,
					Data: pl.Control,
				}
			default:
				log.Error().
					Interface("payload", pl).
					Msg("unknown payload type on data channel")
			}
		case <-p.done:
			break
		}
	}
	log.Info().Str("peer", p.name).Msg("peer run loop returned")
}

func (p *DefaultPeer) RCon(cm *proto.Control) error {
	log.Trace().Stringer("msg", cm).Msg("send control message to peer")

	frame := &proto.Frame{
		Payload: &proto.Frame_Control{cm},
	}

	err := p.data.Send(frame)
	if err != nil {
		return err
	}
	return nil
}

func (p *DefaultPeer) handleICECandidate(c *webrtc.ICECandidate) {
	if c == nil {
		return
	}
	p.signals <- &proto.Frame{
		Dst:     p.name,
		Payload: proto.PayloadWithICECandidate(c.ToJSON()),
	}
}

func (p *DefaultPeer) handleICEStateChange(st webrtc.ICEConnectionState) {
	log.Info().Stringer("state", st).Msg("ICE connection state changed")
	switch st {
	case webrtc.ICEConnectionStateConnected:
		log.Info().Msg("peer connection established")
		p.events <- EventPeerConnected{Peer: p}
		return
	case webrtc.ICEConnectionStateDisconnected:
		p.Close()
		p.events <- EventPeerDisconnected{Peer: p}
	case webrtc.ICEConnectionStateFailed:
	case webrtc.ICEConnectionStateClosed:
	}
}

func (p *DefaultPeer) handleTrack(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	log.Info().Str("track_id", track.ID()).Msg("new track")

	// temporary workaround until pion webrtc implements incoming RTCP events
	go func() {
		ticker := time.NewTicker(time.Second * 3)
		for range ticker.C {
			err := p.conn.WriteRTCP([]rtcp.Packet{
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

	stream, ok := p.str[track.ID()]
	if !ok {
		log.Error().Str("track_id", track.ID()).Msg("stream id unknown")
		return
	}

	log.Trace().Str("track_id", track.ID()).Msg("receive")
	p.events <- EventStreamStart{
		Peer:   p,
		Stream: stream,
	}
	stream.Receive(track)
}

func (p *DefaultPeer) Name() string {
	return p.name
}

// TODO: refactor interface for stream access
func (p *DefaultPeer) VideoLocal() StreamSender {
	return p.stl["video"]
}

func (p *DefaultPeer) AudioLocal() StreamSender {
	return p.stl["audio"]
}

func (p *DefaultPeer) ScreenLocal() StreamSender {
	return p.stl["screen"]
}

func (p *DefaultPeer) VideoRemote() StreamReceiver {
	return p.str["video"]
}

func (p *DefaultPeer) AudioRemote() StreamReceiver {
	return p.str["audio"]
}

func (p *DefaultPeer) ScreenRemote() StreamReceiver {
	return p.str["screen"]
}

func (p *DefaultPeer) Close() {
	for _, stream := range p.str {
		if stream == nil {
			continue
		}
		p.events <- EventStreamEnd{
			Peer:   p,
			Stream: stream,
		}
		stream.Close()
	}
	for _, stream := range p.stl {
		if stream == nil {
			continue
		}
		stream.Close()
	}

	select {
	case <-p.done:
	default:
		p.data.Close()
		p.conn.Close()
		close(p.done)
	}

	log.Info().Str("peer", p.name).Msg("peer connection closed")
}
