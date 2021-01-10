package client

import (
	"fmt"

	"github.com/lx7/devnet/gst"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/rs/zerolog/log"
)

type LocalStreamOpts struct {
	ID     string
	Group  string
	Preset *gst.Preset
}

type StreamSender interface {
	Send()
	Close()
}

type LocalStream struct {
	track    *webrtc.TrackLocalStaticSample
	pipeline *gst.Pipeline
}

func NewLocalStream(c *webrtc.PeerConnection, so *LocalStreamOpts) (*LocalStream, error) {
	t, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: so.Preset.MimeType},
		so.ID,
		so.Group,
	)
	if err != nil {
		return nil, fmt.Errorf("new local track: %v", err)
	}

	if _, err = c.AddTrack(t); err != nil {
		return nil, fmt.Errorf("add local track: %v", err)
	}

	s := &LocalStream{
		track: t,
	}

	log.Debug().Str("pipeline", so.Preset.Local).Msg("new local pipeline")
	s.pipeline, err = gst.NewPipeline(so.Preset.Local)
	if err != nil {
		return nil, fmt.Errorf("new pipeline: %v", err)
	}
	s.pipeline.HandleSample(s.handleSample)

	return s, nil
}

func (s *LocalStream) Send() {
	s.pipeline.Start()
}

func (s *LocalStream) handleSample(sample media.Sample) {
	if err := s.track.WriteSample(sample); err != nil {
		log.Error().Err(err).Msg("write sample to track")
	}
}

func (s *LocalStream) Close() {
	if s == nil || s.pipeline == nil {
		return
	}
	s.pipeline.Stop()
	s.pipeline = nil
	s.track = nil
}
