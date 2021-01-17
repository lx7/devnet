package client

import (
	"testing"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/require"
)

func TestStream_Pipeline(t *testing.T) {
	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	require.NoError(t, err)

	local, err := NewStreamLocal(conn, &StreamOpts{
		ID:    "video",
		Group: "chat",
		Pipeline: `
			videotestsrc 
			! videoconvert 
			! autovideosink
			`,
		MimeType: "video/H264",
	})
	require.NoError(t, err)

	remote, err := NewStreamRemote(conn, StreamOpts{
		ID:    "video",
		Group: "chat",
		Pipeline: `
			videotestsrc 
			! videoconvert 
			! autovideosink
			`,
	})
	require.NoError(t, err)

	conn.OnTrack(func(t *webrtc.TrackRemote, recv *webrtc.RTPReceiver) {
		remote.Receive(t)
	})

	local.Send()
	time.Sleep(2 * time.Second)
}
