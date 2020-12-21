package client

import (
	"runtime"
	"testing"
	"time"

	"github.com/gotk3/gotk3/gtk"
	"github.com/lx7/devnet/internal/testutil"
	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	runtime.LockOSThread()
	//os.Setenv("GST_DEBUG", "*:2")
	//log.SetLevel(log.DebugLevel)
}

func TestSession_Flow(t *testing.T) {
	hook := testutil.NewLogHook()
	gtk.Init(nil)

	conf := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{URLs: []string{
			"stun:stun.l.google.com:19302",
		}}},
	}

	s1, err := NewSession(SessionOpts{
		Peer:  "user2",
		wconf: conf,
	})
	require.NoError(t, err)

	offer, err := s1.CreateOffer()
	require.NoError(t, err)

	s2, err := SessionWithOffer(SessionOpts{
		Peer:  "user1",
		wconf: conf,
	}, offer)
	require.NoError(t, err)

	answer, err := s2.CreateAnswer()
	require.NoError(t, err)

	err = s1.HandleAnswer(answer)
	require.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	//s1.Voice.Start()
	s1.ScreenCast.Start()
	// TODO: improve testing for failure modes caused by gstreamer pipelines

	go func() {
		time.Sleep(10 * time.Second)
		gtk.MainQuit()
	}()

	gtk.Main()

	s1.Close()
	s2.Close()

	errorlog := hook.Entry(log.ErrorLevel)
	if errorlog != nil {
		t.Errorf("runtime error: '%v'", errorlog.Message)
	}
}
