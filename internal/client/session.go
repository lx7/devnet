package client

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/gst"
	"github.com/lx7/devnet/proto"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
)

type SessionOpts struct {
	Self, Peer string
	wconf      webrtc.Configuration

	ScreenCastOverlay *gtk.DrawingArea
}

type Session struct {
	Peer string
	Conn *webrtc.PeerConnection

	Voice      *localStream
	ScreenCast *localStream

	scOverlay *gtk.DrawingArea

	AudioIn *remoteStream
	VideoIn *remoteStream
}

func NewSession(so SessionOpts) (*Session, error) {
	conn, err := webrtc.NewPeerConnection(so.wconf)
	if err != nil {
		return nil, err
	}

	s := Session{
		Peer:      so.Peer,
		Conn:      conn,
		scOverlay: so.ScreenCastOverlay,
	}

	s.Voice, err = LocalStream(s.Conn, &LocalStreamOpts{
		ID:     "devnet-voice",
		Source: "autoaudiosrc",
		Codec:  gst.Opus,
	})
	if err != nil {
		return nil, err
	}

	s.ScreenCast, err = LocalStream(s.Conn, &LocalStreamOpts{
		ID:     "devnet-screen",
		Source: "ximagesrc use-damage=false ! video/x-raw,framerate=25/1 ",
		Codec:  gst.H264,
	})
	if err != nil {
		return nil, err
	}

	s.Conn.OnICEConnectionStateChange(s.handleICEStateChange)
	s.Conn.OnTrack(s.handleTrack)

	_, err = s.Conn.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo)
	if err != nil {
		return nil, err
	}
	_, err = s.Conn.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func SessionWithOffer(so SessionOpts, sd proto.SessionDescription) (*Session, error) {
	if !sd.IsType(proto.SDP_OFFER) {
		return nil, fmt.Errorf("expected sdp type: OFFER")
	}

	s, err := NewSession(so)
	if err != nil {
		return nil, err
	}

	err = s.Conn.SetRemoteDescription(sd.SessionDescription())
	if err != nil {
		return nil, fmt.Errorf("set remote description: %v", err)
	}

	return s, nil
}

func (s *Session) CreateOffer() (*proto.SDP, error) {
	offer, err := s.Conn.CreateOffer(nil)
	if err != nil {
		return nil, fmt.Errorf("create offer: %v", err)
	}
	return proto.SDPWithSD(offer), nil
}

func (s *Session) CreateAnswer() (*proto.SDP, error) {
	answer, err := s.Conn.CreateAnswer(nil)
	if err != nil {
		return nil, fmt.Errorf("create answer: %v", err)
	}

	err = s.Conn.SetLocalDescription(answer)
	if err != nil {
		return nil, fmt.Errorf("set local description: %v", err)
	}
	return proto.SDPWithSD(answer), nil
}

func (s *Session) HandleAnswer(sd proto.SessionDescription) error {
	if !sd.IsType(proto.SDP_ANSWER) {
		return errors.New("expected sdp type: ANSWER")
	}

	err := s.Conn.SetRemoteDescription(sd.SessionDescription())
	if err != nil {
		return fmt.Errorf("set remote description: %v", err)
	}
	return nil
}

func (s *Session) handleICEStateChange(cs webrtc.ICEConnectionState) {
	log.Info("ICE connection state has changed: ", cs.String())
	switch cs {
	case webrtc.ICEConnectionStateConnected:
		return
	case webrtc.ICEConnectionStateDisconnected:
		s.Close()
	case webrtc.ICEConnectionStateFailed:
	case webrtc.ICEConnectionStateClosed:
	}
}

func (s *Session) handleTrack(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
	// temporary workaround until pion webrtc implements incoming RTCP events
	go func() {
		ticker := time.NewTicker(time.Second * 3)
		for range ticker.C {
			err := s.Conn.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{MediaSSRC: track.SSRC()},
			})
			if err == io.ErrClosedPipe {
				return
			} else if err != nil {
				log.Error("rctp send error: ", err)
			}
		}
	}()

	switch kind := track.Kind(); kind {
	case webrtc.RTPCodecTypeAudio:
		var err error
		s.AudioIn, err = RemoteStream(track, RemoteStreamOpts{
			Sink: "autoaudiosink",
		})
		if err != nil {
			return
		}
		s.AudioIn.Receive()
	case webrtc.RTPCodecTypeVideo:
		var err error
		s.VideoIn, err = RemoteStream(track, RemoteStreamOpts{
			Sink:    "autovideosink",
			Overlay: s.scOverlay,
		})
		if err != nil {
			return
		}
		s.VideoIn.Receive()
	default:
		log.Errorf("track of unkown kind: %d", track.Kind())
		return
	}

}

func (s *Session) Close() {
	log.Info("session closed")

	s.Voice.Close()
	s.ScreenCast.Close()

	s.AudioIn.Close()
	s.VideoIn.Close()

	s.Conn.Close()
}
