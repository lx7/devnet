package client

import (
	"fmt"
	"io"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/gst"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
)

type RemoteStreamOpts struct {
	ID     string
	Group  string
	Preset *gst.Preset
}

type StreamReceiver interface {
	SetOverlay(gtk.IWidget) error
	Receive(*webrtc.TrackRemote)
	Close()
}

type RemoteStream struct {
	track    *webrtc.TrackRemote
	pipeline *gst.Pipeline
}

func NewRemoteStream(c *webrtc.PeerConnection, so RemoteStreamOpts) (*RemoteStream, error) {
	s := &RemoteStream{}

	log.Debug().Str("pipeline", so.Preset.Remote).Msg("new remote pipeline")
	p, err := gst.NewPipeline(so.Preset.Remote)
	if err != nil {
		return nil, fmt.Errorf("new pipeline: %v", err)
	}
	s.pipeline = p

	return s, nil
}

func (s *RemoteStream) SetOverlay(w gtk.IWidget) error {
	return s.pipeline.SetOverlayHandle(w)
}

func (s *RemoteStream) Receive(t *webrtc.TrackRemote) {
	s.track = t
	s.pipeline.Start()

	buf := make([]byte, 1400)
	for {
		i, _, err := t.Read(buf)
		if err == io.EOF {
			s.pipeline.Stop()
			return
		} else if err != nil {
			log.Error().Err(err).Msg("reading track buffer")
		}
		s.pipeline.Push(buf[:i])
	}
}

func (s *RemoteStream) Close() {
	if s == nil || s.pipeline == nil {
		return
	}
	//s.pipeline.Destroy()
	s.track = nil
}
