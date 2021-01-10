// +build darwin

package gst

var screen_H264_SW = Preset{
	MimeType: MimeTypeVideoH264,
	Codec:    H264,
	HW:       NoHardware,
	Source:   Screen,
	Local: `
			avfvideosrc capture-screen=true 
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
}

var screen_H264_OSXVT = Preset{
	MimeType: MimeTypeVideoH264,
	Codec:    H264,
	HW:       OSXVT,
	Source:   Screen,
	Local: `
			avfvideosrc capture-screen=true 
			! video/x-raw,framerate=25/1
			! videoscale
			! videoconvert
			! queue
			! vtenc_h264
			! video/x-h264,stream-format=byte-stream,profile=high 
			! appsink name=sink
			`,
	Remote: `
			appsrc name=src format=time is-live=true do-timestamp=true
			! application/x-rtp
			! rtph264depay 
    		! queue 
			! vtdec_h264
			! videoconvert 
			! autovideosink sync=false
			`,
}

var voice_OPUS_SW = Preset{
	MimeType:   MimeTypeAudioOpus,
	Codec:      Opus,
	HW:         NoHardware,
	SourceType: Voice,
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
}

// presets holds the list of presets that are enabled for this platform
var presets = []Preset{
	screen_H264_SW,
	screen_H264_OSXVT,
	voice_OPUS_SW,
}
