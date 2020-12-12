package gst

import (
	"fmt"

	"github.com/pion/webrtc/v2"
)

type codec struct {
	Enc   string
	Dec   string
	Clock int
}

/*
source:
    ximagesrc use-damage=false
    ! video/x-raw,framerate=30/1
*/

// recv: "appsrc name=receive format=time is-live=true do-timestamp=true"
// send: "appsink name=send"

var codecs = map[string]codec{
	webrtc.H264: codec{
		Enc: `
			! video/x-raw,format=I420 
			! x264enc 
				tune=zerolatency 
				key-int-max=120 
				speed-preset=ultrafast 
			! video/x-h264,stream-format=byte-stream 
			`,
		Dec: `
			application/x-rtp
			! rtph264depay 
    		! queue 
			! decodebin 
			! autovideosink"
			`,
		Clock: 90000,
	},
	webrtc.VP8: codec{
		Enc: `
			! vp8enc 
				error-resilient=partitions 
				keyframe-max-dist=10 
				auto-alt-ref=true 
				cpu-used=5 
				deadline=1
			`,
		Dec: `
			application/x-rtp, encoding-name=VP8-DRAFT-IETF-01
			! rtpvp8depay 
			! decodebin 
			! autovideosink
			`,
		Clock: 90000,
	},
	webrtc.VP9: codec{
		Enc: `
			! vp9enc 
			`,
		Dec: `
			application/x-rtp
			! rtpvp9depay 
			! decodebin 
			! autovideosink"
			`,
		Clock: 90000,
	},
	webrtc.Opus: codec{
		Enc: `
			! opusenc
			`,
		Dec: `
			application/x-rtp, payload=96, encoding-name=OPUS
			! rtpopusdepay 
			! decodebin 
			! autoaudiosink"
			`,
		Clock: 48000,
	},
}

func Codec(name string) (c codec, err error) {
	c, ok := codecs[name]
	if !ok {
		return c, fmt.Errorf("unknown codec: %v", name)
	}
	return c, nil
}
