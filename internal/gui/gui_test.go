package gui

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/gst"
	"github.com/lx7/devnet/internal/client"
	"github.com/lx7/devnet/proto"
	"github.com/pion/webrtc/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	runtime.LockOSThread()
}

func TestGUI_Events(t *testing.T) {
	time.Sleep(1 * time.Second)

	s := newFakeSession()

	// set expectations
	s.On("Events").Return()

	gui, err := New("test.devnet.events", s)
	assert.NoError(t, err, "constructor should not fail")

	peer := newFakePeer("peer1")

	// define test cases
	tests := []struct {
		desc  string
		give  client.Event
		check func(*testing.T)
	}{
		{
			desc: "signaling connected",
			give: client.EventConnected{},
			check: func(t *testing.T) {
				assert.Equal(t, false, gui.mainWindow.waitScreen.IsVisible())
				assert.Equal(t, true, gui.mainWindow.channelList.IsVisible())
				assert.Equal(t, false, gui.mainWindow.detailsBox.IsVisible())
			},
		},
		{
			desc: "peer connected",
			give: client.EventPeerConnected{Peer: peer},
			check: func(t *testing.T) {
				assert.Equal(t, true, gui.mainWindow.detailsBox.IsVisible())
			},
		},
		{
			desc: "remote video start",
			give: client.EventStreamStart{Peer: peer, Stream: peer.videoRemote},
			check: func(t *testing.T) {
				peer.videoRemote.pipeline.Start()
				assert.Equal(t, false, gui.videoWindow.IsVisible())
			},
		},
		{
			desc: "remote screen start",
			give: client.EventStreamStart{Peer: peer, Stream: peer.screenRemote},
			check: func(t *testing.T) {
				peer.screenRemote.pipeline.Start()
				assert.Equal(t, true, gui.videoWindow.IsVisible())
			},
		},
		{
			desc: "remote screen end",
			give: client.EventStreamEnd{Peer: peer, Stream: peer.screenRemote},
			check: func(t *testing.T) {
				assert.Equal(t, false, gui.videoWindow.IsVisible())
			},
		},
		{
			desc: "peer diconnected",
			give: client.EventPeerDisconnected{Peer: peer},
			check: func(t *testing.T) {
				assert.Equal(t, false, gui.mainWindow.detailsBox.IsVisible())
			},
		},
	}

	// run tests
	go func() {
		const interval = 1000 * time.Millisecond
		for _, tt := range tests {
			time.Sleep(interval)
			s.events <- tt.give
			time.Sleep(10 * time.Millisecond)
			t.Run(tt.desc, tt.check)
		}
		time.Sleep(interval)
		glib.IdleAdd(gui.Quit)
	}()

	exitcode := gui.Run()
	assert.Equal(t, 0, exitcode, "gui should exit with code 0")
	time.Sleep(1 * time.Second)
}

type fakeLocalStream struct {
	pipeline *gst.Pipeline
	id       string
}

func (s *fakeLocalStream) SetOverlay(o gtk.IWidget) error {
	s.pipeline.SetOverlayHandle(o)
	return nil
}

func (s *fakeLocalStream) ID() string {
	return s.id
}

func (s *fakeLocalStream) Send() {
	s.pipeline.Start()
}

func (s *fakeLocalStream) Stop() {
	s.pipeline.Stop()
}

func (s *fakeLocalStream) Close() {
}

type fakeRemoteStream struct {
	pipeline *gst.Pipeline
	id       string
}

func (s *fakeRemoteStream) SetOverlay(o gtk.IWidget) error {
	s.pipeline.SetOverlayHandle(o)
	return nil
}

func (s *fakeRemoteStream) ID() string {
	return s.id
}

func (s *fakeRemoteStream) Receive(*webrtc.TrackRemote) {
}

func (s *fakeRemoteStream) Close() {
}

func (s *fakeRemoteStream) Pause() {
}

func (s *fakeRemoteStream) Unpause() {
}

type fakePeer struct {
	name string

	videoLocal   *fakeLocalStream
	videoRemote  *fakeRemoteStream
	audioLocal   *fakeLocalStream
	audioRemote  *fakeRemoteStream
	screenLocal  *fakeLocalStream
	screenRemote *fakeRemoteStream
}

func newFakePeer(name string) *fakePeer {
	p := &fakePeer{name: name}

	pr1, _ := gst.NewPipeline("videotestsrc ! autovideosink")
	p.screenRemote = &fakeRemoteStream{
		pipeline: pr1,
		id:       "screen",
	}

	pr2, _ := gst.NewPipeline("videotestsrc pattern=snow ! autovideosink")
	p.videoRemote = &fakeRemoteStream{
		pipeline: pr2,
		id:       "video",
	}

	pr3, _ := gst.NewPipeline("audiotestsrc ! autoaudiosink")
	p.audioRemote = &fakeRemoteStream{
		pipeline: pr3,
		id:       "audio",
	}

	pl1, _ := gst.NewPipeline("videotestsrc ! autovideosink")
	p.screenLocal = &fakeLocalStream{
		pipeline: pl1,
	}

	pl2, _ := gst.NewPipeline("videotestsrc pattern=ball ! autovideosink")
	p.videoLocal = &fakeLocalStream{
		pipeline: pl2,
	}

	pl3, _ := gst.NewPipeline("audiotestsrc ! fakesink")
	p.audioLocal = &fakeLocalStream{
		pipeline: pl3,
	}

	return p
}

func (p *fakePeer) VideoLocal() client.StreamSender {
	return p.videoLocal
}
func (p *fakePeer) VideoRemote() client.StreamReceiver {
	return p.videoRemote
}

func (p *fakePeer) AudioLocal() client.StreamSender {
	return p.audioLocal
}
func (p *fakePeer) AudioRemote() client.StreamReceiver {
	return p.audioRemote
}

func (p *fakePeer) ScreenLocal() client.StreamSender {
	return p.screenLocal
}
func (p *fakePeer) ScreenRemote() client.StreamReceiver {
	return p.screenRemote
}

func (p *fakePeer) HandleSignaling(*proto.Frame) error {
	return nil
}

func (p *fakePeer) Name() string {
	return p.name
}

func (p *fakePeer) Close() {
}

type fakeSession struct {
	mock.Mock
	peers  []*fakePeer
	events chan client.Event
}

func newFakeSession() *fakeSession {
	s := &fakeSession{
		peers: []*fakePeer{
			newFakePeer("peer1"),
		},
		events: make(chan client.Event, 5),
	}
	return s
}

func (s *fakeSession) Events() <-chan client.Event {
	s.Called()
	return s.events
}

func (s *fakeSession) Connect(peer string) error {
	s.Called(peer)
	return nil
}
