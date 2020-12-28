// +build linux

package gst

import (
	"github.com/pion/webrtc/v2"
)

var screen_H264_SW = Preset{
	Kind:   Video,
	Codec:  H264,
	HW:     NoHardware,
	Source: Screen,
	Local: `
			ximagesrc use-damage=false 
			! video/x-raw,framerate=25/1
			! videoscale
			! videoconvert
			! queue
			! x264enc 
				tune=zerolatency 
				key-int-max=60 
				speed-preset=ultrafast 
			! video/x-h264,stream-format=byte-stream,profile=high 
			! appsink name=sink
			`,
	Remote: `
			appsrc name=src format=time is-live=true do-timestamp=true
			! application/x-rtp
			! rtph264depay 
    		! queue 
			! decodebin
			! videoconvert 
			! autovideosink sync=false
			`,
	Clock:       ClockRateVideo,
	PayloadType: webrtc.DefaultPayloadTypeH264,
}

var screen_H264_VAAPI = Preset{
	Kind:   Video,
	Codec:  H264,
	HW:     VAAPI,
	Source: Screen,
	Local: `
			ximagesrc use-damage=false 
			! video/x-raw,framerate=25/1
			! vaapipostproc
			! queue
			! vaapih264enc 
				cpb-length=300
				quality-level=7
				keyframe-period=0
				compliance-mode=1
				cabac=1
			! video/x-h264,stream-format=byte-stream,profile=high
			! appsink name=sink
			`,
	Remote: `
			appsrc name=src format=time is-live=true do-timestamp=true
			! application/x-rtp
			! rtph264depay 
    		! h264parse
			! vaapih264dec low-latency=true
			! queue
			! vaapipostproc
			! vaapisink sync=false
			`,
	Clock:       ClockRateVideo,
	PayloadType: webrtc.DefaultPayloadTypeH264,
}

var screen_H264_NVCODEC = Preset{
	Kind:   Video,
	Codec:  H264,
	HW:     NVCODEC,
	Source: Screen,
	Local: `
			ximagesrc use-damage=false 
			! video/x-raw,framerate=25/1
			! videoconvert
			! queue
			! nvh264enc 
				preset=low-latency
			! video/x-h264,stream-format=byte-stream,profile=high
			! appsink name=sink
			`,
	Remote: `
			appsrc name=src format=time is-live=true do-timestamp=true
			! application/x-rtp
			! rtph264depay 
			! decodebin 
			! glimagesink sync=false
			`,
	Clock:       ClockRateVideo,
	PayloadType: webrtc.DefaultPayloadTypeH264,
}

var voice_OPUS_SW = Preset{
	Kind:   Audio,
	Codec:  Opus,
	HW:     NoHardware,
	Source: Voice,
	Local: `
			autoaudiosrc
			! opusenc
			! appsink name=sink
			`,
	Remote: `
			appsrc name=src format=time is-live=true do-timestamp=true
			! application/x-rtp, payload=96, encoding-name=OPUS
			! rtpopusdepay 
			! decodebin 
			! queue
			! autoaudiosink
			`,
	Clock:       ClockRateVideo,
	PayloadType: webrtc.DefaultPayloadTypeOpus,
}

// presets holds the list of presets that are enabled for this platform
var presets = []Preset{
	screen_H264_SW,
	screen_H264_VAAPI,
	//screen_H264_NVCODEC,
	voice_OPUS_SW,
}
