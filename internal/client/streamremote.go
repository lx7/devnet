package client

import (
	"fmt"
	"io"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/gst"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog/log"
)

const (
	inactive_timeout = 800 * time.Millisecond
	inactive_ticker  = 100 * time.Millisecond
)

type StreamRemote struct {
	track     *webrtc.TrackRemote
	pipeline  *gst.Pipeline
	lastframe time.Time
	active    bool
}

func NewStreamRemote(c *webrtc.PeerConnection, so StreamOpts) (*StreamRemote, error) {
	s := &StreamRemote{}

	log.Debug().Str("pipeline", so.Pipeline).Msg("new remote pipeline")
	p, err := gst.NewPipeline(so.Pipeline)
	if err != nil {
		return nil, fmt.Errorf("new pipeline: %v", err)
	}
	s.pipeline = p

	return s, nil
}

func (s *StreamRemote) ID() string {
	return s.track.ID()
}

func (s *StreamRemote) SetOverlay(w gtk.IWidget) error {
	return s.pipeline.SetOverlayHandle(w)
}

func (s *StreamRemote) Receive(t *webrtc.TrackRemote) {
	s.track = t
	s.pipeline.Start()
	s.active = true

	ticker := time.NewTicker(inactive_timeout)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			if s.track == nil || s.pipeline == nil {
				return
			}
			if time.Since(s.lastframe) > inactive_timeout {
				s.pipeline.Pause()
				s.active = false
			}
		}
	}()

	buf := make([]byte, 1400)
	for {
		i, _, err := t.Read(buf)
		if err == io.EOF {
			s.pipeline.Stop()
			return
		} else if err != nil {
			log.Error().Err(err).Msg("reading track buffer")
		}

		if !s.active {
			s.pipeline.Start()
			s.active = true
		}
		s.pipeline.Push(buf[:i])
		s.lastframe = time.Now()
	}
}

func (s *StreamRemote) Close() {
	if s == nil || s.pipeline == nil {
		return
	}
	//s.pipeline.Destroy()
	s.track = nil
}
