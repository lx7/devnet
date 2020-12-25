package gui

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/gst"
	"github.com/lx7/devnet/internal/client"
	"github.com/pion/webrtc/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	runtime.LockOSThread()
}

func TestGUI_Session(t *testing.T) {
	sChan := make(chan *fakeSession)

	// run session in a separate thread
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := newFakeSession()
		s.events <- client.EventSessionStart{}

		// define expectations
		s.On("Events").Return()
		s.On("SetOverlay", client.RemoteScreen, mock.Anything).Return()
		//s.On("StartStream", client.LocalScreen).Return()
		//s.On("Connect").Return()

		sChan <- s

		const interval = 1000 * time.Millisecond
		time.Sleep(1 * time.Second)

		s.testStartInbound()
		time.Sleep(interval)

		s.testStopInbound()
		time.Sleep(interval)

		s.events <- client.EventSessionEnd{}
		s.AssertExpectations(t)
	}()

	gui, err := New("test.devnet", <-sChan)
	assert.NoError(t, err, "constructor should not fail")

	go func() {
		wg.Wait()
		const interval = 1000 * time.Millisecond
		glib.IdleAdd(gui.Quit)
	}()

	exitcode := gui.Run()
	assert.Equal(t, 0, exitcode, "gui should exit with code 0")
}

func TestGUI_Interface(t *testing.T) {
	time.Sleep(1 * time.Second)

	sChan := make(chan *fakeSession)
	// run session in separate thread
	go func() {
		s := newFakeSession()
		s.On("Events").Return()
		s.On("SetOverlay", mock.Anything, mock.Anything).Return()
		s.On("StartStream", mock.Anything).Return()

		s.events <- client.EventSessionStart{}
		sChan <- s
	}()

	gui, err := New("test.devnet", <-sChan)
	assert.NoError(t, err, "constructor should not fail")

	// run tests
	go func() {
		const interval = 1000 * time.Millisecond
		time.Sleep(interval)

		glib.IdleAdd(gui.mainWindow.shareButton.SetActive, true)
		time.Sleep(interval)

		glib.IdleAdd(gui.mainWindow.shareButton.SetActive, true)
		time.Sleep(interval)

		glib.IdleAdd(gui.mainWindow.shareButton.SetActive, false)
		time.Sleep(interval)

		glib.IdleAdd(gui.Quit)
	}()

	exitcode := gui.Run()
	assert.Equal(t, 0, exitcode, "gui should exit with code 0")
}

type fakeLocalStream struct {
}

func (s *fakeLocalStream) Send() {
}

func (s *fakeLocalStream) Close() {
}

type fakeRemoteStream struct {
	pipeline *gst.Pipeline
}

func (s *fakeRemoteStream) SetOverlay(o gtk.IWidget) error {
	s.pipeline.SetOverlayHandle(o)
	return nil
}

func (s *fakeRemoteStream) Receive(*webrtc.Track) {
}

func (s *fakeRemoteStream) Close() {
}

type fakeSession struct {
	mock.Mock
	screen   *fakeLocalStream
	screenIn *fakeRemoteStream
	events   chan client.Event
}

func newFakeSession() *fakeSession {
	fakeRemotePipe, _ := gst.NewPipeline("videotestsrc ! autovideosink", 90000)
	s := &fakeSession{
		screen: &fakeLocalStream{},
		screenIn: &fakeRemoteStream{
			pipeline: fakeRemotePipe,
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

func (s *fakeSession) SetOverlay(id int, o *gtk.DrawingArea) {
	s.Called(id, o)
	s.screenIn.pipeline.SetOverlayHandle(o)
}

func (s *fakeSession) StartStream(id int) {
	s.Called(id)
	s.screen.Send()
}

func (s *fakeSession) testStartInbound() {
	s.screenIn.pipeline.Start()
	s.events <- client.EventSCInboundStart{}
}

func (s *fakeSession) testStopInbound() {
	s.screenIn.pipeline.Stop()
	s.events <- client.EventSCInboundEnd{}
}
