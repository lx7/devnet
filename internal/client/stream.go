package client

import (
	"github.com/gotk3/gotk3/gtk"
	"github.com/pion/webrtc/v3"
)

type StreamOpts struct {
	ID       string
	Group    string
	Pipeline string
	MimeType string
}

type Stream interface {
	ID() string
	SetOverlay(gtk.IWidget) error
	Close()
}

type StreamReceiver interface {
	Stream
	Receive(*webrtc.TrackRemote)
}

type StreamSender interface {
	Stream
	Send()
	Stop()
}
