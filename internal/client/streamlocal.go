package client

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/gst"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/rs/zerolog/log"
)

type StreamLocal struct {
	id       string
	track    *webrtc.TrackLocalStaticSample
	pipeline *gst.Pipeline
}

func NewStreamLocal(c *webrtc.PeerConnection, so *StreamOpts) (*StreamLocal, error) {
	t, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: so.MimeType},
		so.ID,
		so.Group,
	)
	if err != nil {
		return nil, fmt.Errorf("new local track: %v", err)
	}

	if _, err = c.AddTrack(t); err != nil {
		return nil, fmt.Errorf("add local track: %v", err)
	}

	s := &StreamLocal{
		id:    so.ID,
		track: t,
	}

	log.Debug().Str("pipeline", so.Pipeline).Msg("new local pipeline")
	s.pipeline, err = gst.NewPipeline(so.Pipeline)
	if err != nil {
		return nil, fmt.Errorf("new pipeline: %v", err)
	}
	s.pipeline.HandleSample(s.handleSample)

	return s, nil
}

func (s *StreamLocal) ID() string {
	return s.id
}

func (s *StreamLocal) SetOverlay(w gtk.IWidget) error {
	return s.pipeline.SetOverlayHandle(w)
}

func (s *StreamLocal) Send() {
	if s.pipeline == nil {
		return
	}
	s.pipeline.Start()
}

func (s *StreamLocal) Stop() {
	if s.pipeline == nil {
		return
	}
	s.pipeline.Stop()
}

func (s *StreamLocal) handleSample(sample media.Sample) {
	if err := s.track.WriteSample(sample); err != nil {
		log.Error().Err(err).Msg("write sample to track")
	}
}

func (s *StreamLocal) Close() {
	if s == nil || s.pipeline == nil {
		return
	}
	s.pipeline.Stop()
	s.pipeline = nil
	s.track = nil
}
