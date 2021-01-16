package gui

import (
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/gst"
	"github.com/lx7/devnet/internal/client"
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
		s.On("SetOverlay", client.RemoteCamera, mock.Anything).Return()
		s.On("SetOverlay", client.RemoteScreen, mock.Anything).Return()
		s.On("SetOverlay", client.LocalCamera, mock.Anything).Return()
		//s.On("StartStream", client.LocalScreen).Return()
		//s.On("Connect").Return()

		sChan <- s

		const interval = 1000 * time.Millisecond
		time.Sleep(1 * time.Second)

		s.testStartRemoteCam()
		time.Sleep(1 * interval)

		s.testStopRemoteCam()
		time.Sleep(interval)

		s.testStartRemoteScreen()
		time.Sleep(1 * interval)

		s.testStopRemoteScreen()
		time.Sleep(interval)

		s.events <- client.EventSessionEnd{}
		s.AssertExpectations(t)
	}()

	gui, err := New("test.devnet", <-sChan)
	assert.NoError(t, err, "constructor should not fail")

	go func() {
		wg.Wait()
		glib.IdleAdd(gui.Quit)
	}()

	exitcode := gui.Run()
	assert.Equal(t, 0, exitcode, "gui should exit with code 0")
}

func TestGUI_Interface(t *testing.T) {
	time.Sleep(1 * time.Second)

	s := newFakeSession()
	s.On("Events").Return()
	s.On("SetOverlay", mock.Anything, mock.Anything).Return()
	s.On("StartStream", mock.Anything).Return()
	s.On("StopStream", mock.Anything).Return()

	gui, err := New("test.devnet", s)
	assert.NoError(t, err, "constructor should not fail")

	// run tests
	go func() {
		const interval = 1000 * time.Millisecond
		time.Sleep(interval)

		s.events <- client.EventConnected{}
		time.Sleep(interval)

		s.events <- client.EventSessionStart{}
		time.Sleep(interval)

		s.events <- client.EventDisconnected{}
		time.Sleep(interval)

		s.events <- client.EventConnected{}
		time.Sleep(interval)

		s.events <- client.EventSCInboundStart{}
		time.Sleep(interval)

		s.events <- client.EventSCInboundEnd{}
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
	pipeline *gst.Pipeline
}

func (s *fakeLocalStream) SetOverlay(o gtk.IWidget) error {
	s.pipeline.SetOverlayHandle(o)
	return nil
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
}

func (s *fakeRemoteStream) SetOverlay(o gtk.IWidget) error {
	s.pipeline.SetOverlayHandle(o)
	return nil
}

func (s *fakeRemoteStream) Receive(*webrtc.TrackRemote) {
}

func (s *fakeRemoteStream) Close() {
}

type fakeSession struct {
	mock.Mock
	streams []client.Stream
	events  chan client.Event
}

func newFakeSession() *fakeSession {
	s := &fakeSession{
		streams: make([]client.Stream, 10),
		events:  make(chan client.Event, 5),
	}

	pr1, _ := gst.NewPipeline("videotestsrc ! autovideosink")
	pr2, _ := gst.NewPipeline("videotestsrc pattern=snow ! autovideosink")

	s.streams[client.RemoteScreen] = &fakeRemoteStream{
		pipeline: pr1,
	}
	s.streams[client.RemoteCamera] = &fakeRemoteStream{
		pipeline: pr2,
	}

	pl1, _ := gst.NewPipeline("videotestsrc ! autovideosink")
	pl2, _ := gst.NewPipeline("videotestsrc pattern=ball ! autovideosink")

	s.streams[client.LocalScreen] = &fakeLocalStream{
		pipeline: pl1,
	}
	s.streams[client.LocalCamera] = &fakeLocalStream{
		pipeline: pl2,
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

func (s *fakeSession) SetOverlay(id int, o *gtk.GLArea) {
	s.Called(id, o)
	s.streams[id].SetOverlay(o)
}

func (s *fakeSession) StartStream(id int) error {
	//s.Called(id)
	stream, _ := s.streams[id].(client.StreamSender)
	stream.Send()
	return nil
}

func (s *fakeSession) StopStream(id int) error {
	//s.Called(id)
	stream, _ := s.streams[id].(client.StreamSender)
	stream.Stop()
	return nil
}

func (s *fakeSession) testStartRemoteCam() {
	stream, _ := s.streams[client.RemoteCamera].(*fakeRemoteStream)
	stream.pipeline.Start()
	s.events <- client.EventCameraInboundStart{}
}

func (s *fakeSession) testStopRemoteCam() {
	stream, _ := s.streams[client.RemoteCamera].(*fakeRemoteStream)
	stream.pipeline.Stop()
	s.events <- client.EventCameraInboundEnd{}
}

func (s *fakeSession) testStartRemoteScreen() {
	stream, _ := s.streams[client.RemoteScreen].(*fakeRemoteStream)
	stream.pipeline.Start()
	s.events <- client.EventSCInboundStart{}
}

func (s *fakeSession) testStopRemoteScreen() {
	stream, _ := s.streams[client.RemoteScreen].(*fakeRemoteStream)
	stream.pipeline.Stop()
	s.events <- client.EventSCInboundEnd{}
}
