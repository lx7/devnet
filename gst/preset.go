package gst

import (
	"fmt"
	"strings"
)

const (
	MimeTypeVideoVP8  = "video/vp8"
	MimeTypeVideoH264 = "video/h264"
	MimeTypeAudioOpus = "audio/opus"
)

// TODO: switch constants to int?
type codecName string

const (
	H264 codecName = "H264"
	Opus codecName = "Opus"
)

type hwCodec string

const (
	NoHardware hwCodec = ""
	VAAPI      hwCodec = "vaapi"
	NVCODEC    hwCodec = "nvcodec"
	VDPAU      hwCodec = "vdpau"
	OSXVT      hwCodec = "osxvt"
)

type sourceType string

const (
	Screen sourceType = "screen"
	Camera sourceType = "camera"
	Voice  sourceType = "voice"
)

type Preset struct {
	MimeType string
	Codec    codecName
	HW       hwCodec
	Source   sourceType
	Local    string
	Remote   string
}

func NewHardwareCodec(s string) hwCodec {
	switch strings.ToLower(s) {
	case string(VAAPI):
		return VAAPI
	case string(NVCODEC):
		return NVCODEC
	case string(VDPAU):
		return VDPAU
	case string(OSXVT):
		return OSXVT
	default:
		return NoHardware
	}
}

func (c *Preset) String() string {
	return fmt.Sprintf("%s/%s (%s)", c.Codec, c.HW, c.Source)
}

func GetPreset(s sourceType, c codecName, h hwCodec) (*Preset, error) {
	for _, p := range presets {
		if p.Source == s && p.Codec == c && p.HW == h {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("preset %s/%s (%s) not found", c, h, s)
}

func PresetsBySource(s sourceType) []Preset {
	var ps []Preset
	for _, p := range presets {
		if p.Source != s {
			continue
		}
		ps = append(ps, p)
	}
	return ps
}
