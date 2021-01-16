package client

import (
	"github.com/gotk3/gotk3/gtk"
	"github.com/pion/webrtc/v3"
)

type Stream interface {
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
