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
	Source string
	Codec  gst.Codec
}

type localStream struct {
	track    *webrtc.Track
	codec    gst.Codec
	pipeline *gst.Pipeline
}

const outboundAppSink = "appsink name=sink"

func LocalStream(c *webrtc.PeerConnection, so *LocalStreamOpts) (*localStream, error) {
	name := fmt.Sprintf("%s-%s-out", so.Codec.Kind, so.Codec.Name)

	t, err := c.NewTrack(so.Codec.PayloadType, rand.Uint32(), name, so.ID)
	if err != nil {
		return nil, fmt.Errorf("new local track: %v", err)
	}

	if _, err = c.AddTrack(t); err != nil {
		return nil, fmt.Errorf("add local track: %v", err)
	}

	s := &localStream{
		codec: so.Codec,
		track: t,
	}

	descr := fmt.Sprintf("%s %s! %s", so.Source, so.Codec.Enc, outboundAppSink)
	log.Debugf("new local pipeline: %s", descr)

	s.pipeline, err = gst.NewPipeline(descr, s.codec.Clock)
	if err != nil {
		return nil, fmt.Errorf("new pipeline: %v", err)
	}
	s.pipeline.HandleSample(s.handleSample)

	return s, nil
}

func (s *localStream) Start() {
	s.pipeline.Start()
}

func (s *localStream) Close() {
	if s == nil || s.pipeline == nil {
		return
	}
	s.pipeline.Stop()
	s.track = nil
}

func (s *localStream) handleSample(sample media.Sample) {
	if err := s.track.WriteSample(sample); err != nil {
		log.Errorf("write sample to track: %v", err)
	}
}

type RemoteStreamOpts struct {
	Sink    string
	Overlay *gtk.DrawingArea
}

type remoteStream struct {
	track    *webrtc.Track
	codec    gst.Codec
	pipeline *gst.Pipeline
}

const inboundAppSrc = "appsrc name=src format=time is-live=true do-timestamp=true"

func RemoteStream(t *webrtc.Track, so RemoteStreamOpts) (*remoteStream, error) {
	codec, err := gst.CodecByName(t.Codec().Name)
	if err != nil {
		return nil, fmt.Errorf("codec for new track: %v", err)
	}
	s := &remoteStream{
		codec: codec,
		track: t,
	}

	descr := fmt.Sprintf("%s %s! %s", inboundAppSrc, codec.Dec, so.Sink)
	log.Debugf("new inbound pipeline: %s", descr)
	s.pipeline, err = gst.NewPipeline(descr, codec.Clock)
	if err != nil {
		return nil, fmt.Errorf("new pipeline: %v", err)
	}
	if so.Overlay != nil {
		s.pipeline.SetOverlayHandle(so.Overlay)
	}

	return s, nil
}

func (s *remoteStream) SetOverlayHandle(w gtk.IWidget) error {
	return s.pipeline.SetOverlayHandle(w)
}

func (s *remoteStream) Receive() {
	s.pipeline.Start()
	buf := make([]byte, 1400)
	for {
		i, err := s.track.Read(buf)
		if err == io.EOF {
			s.pipeline.Stop()
			return
		} else if err != nil {
			log.Error("reading track buffer: ", err)
		}
		s.pipeline.Push(buf[:i])
	}
}

func (s *remoteStream) Close() {
	if s == nil || s.pipeline == nil {
		return
	}
	s.pipeline.Stop()
	s.track = nil
}
