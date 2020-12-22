// +build dragonfly freebsd netbsd openbsd linux

package gst

import (
	"github.com/pion/webrtc/v2"
)

var screen_H264_SW = Preset{
	CodecKind:  VIDEO,
	CodecName:  H264,
	Accel:      AccelTypeNone,
	SourceType: SourceTypeScreen,
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
	CodecKind:  VIDEO,
	CodecName:  H264,
	Accel:      AccelTypeVAAPI,
	SourceType: SourceTypeScreen,
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

var voice_OPUS_SW = Preset{
	CodecKind:  AUDIO,
	CodecName:  Opus,
	Accel:      AccelTypeNone,
	SourceType: SourceTypeVoice,
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
			! autoaudiosink
			`,
	Clock:       ClockRateVideo,
	PayloadType: webrtc.DefaultPayloadTypeOpus,
}

// presets holds the list of presets that are enabled for this platform
var presets = []Preset{
	screen_H264_SW,
	screen_H264_VAAPI,
	voice_OPUS_SW,
}
