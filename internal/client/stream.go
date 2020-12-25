package client

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/gst"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	log "github.com/sirupsen/logrus"
)

type LocalStreamOpts struct {
	ID     string
	Preset *gst.Preset
}

type StreamSender interface {
	Send()
	Close()
}

type LocalStream struct {
	track    *webrtc.Track
	pipeline *gst.Pipeline
}

func NewLocalStream(c *webrtc.PeerConnection, so *LocalStreamOpts) (*LocalStream, error) {
	name := fmt.Sprintf("%s-%s-out", so.Preset.Kind, so.Preset.Codec)

	t, err := c.NewTrack(so.Preset.PayloadType, rand.Uint32(), name, so.ID)
	if err != nil {
		return nil, fmt.Errorf("new local track: %v", err)
	}

	if _, err = c.AddTrack(t); err != nil {
		return nil, fmt.Errorf("add local track: %v", err)
	}

	s := &LocalStream{
		track: t,
	}

	log.Debugf("new local pipeline: %s", so.Preset.Local)
	s.pipeline, err = gst.NewPipeline(so.Preset.Local, so.Preset.Clock)
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
		log.Errorf("write sample to track: %v", err)
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

type RemoteStreamOpts struct {
	ID     string
	Preset *gst.Preset
}

type StreamReceiver interface {
	SetOverlay(gtk.IWidget) error
	Receive(*webrtc.Track)
	Close()
}

type RemoteStream struct {
	track    *webrtc.Track
	pipeline *gst.Pipeline
}

func NewRemoteStream(c *webrtc.PeerConnection, so RemoteStreamOpts) (*RemoteStream, error) {
	s := &RemoteStream{}

	_, err := c.AddTransceiverFromKind(webrtc.NewRTPCodecType(string(so.Preset.Kind)))
	if err != nil {
		return nil, err
	}

	log.Debugf("new inbound pipeline: %s", so.Preset.Remote)
	p, err := gst.NewPipeline(so.Preset.Remote, so.Preset.Clock)
	if err != nil {
		return nil, fmt.Errorf("new pipeline: %v", err)
	}
	s.pipeline = p

	return s, nil
}

func (s *RemoteStream) SetOverlay(w gtk.IWidget) error {
	return s.pipeline.SetOverlayHandle(w)
}

func (s *RemoteStream) Receive(t *webrtc.Track) {
	s.track = t
	s.pipeline.Start()

	buf := make([]byte, 1400)
	for {
		i, err := t.Read(buf)
		if err == io.EOF {
			s.pipeline.Stop()
			return
		} else if err != nil {
			log.Error("reading track buffer: ", err)
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
