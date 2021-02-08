package gst

import (
	"fmt"
	"strings"
)

const (
	MimeTypeVP8  = "video/vp8"
	MimeTypeH264 = "video/h264"
	MimeTypeOpus = "audio/opus"
)

type hwCodec string

const (
	NoHardware hwCodec = ""
	Auto       hwCodec = "auto"
	VAAPI      hwCodec = "vaapi"
	NVCODEC    hwCodec = "nvcodec"
	VDPAU      hwCodec = "vdpau"
	OSXVT      hwCodec = "osxvt"
)

const (
	Screen = "screen"
	Camera = "camera"
	Voice  = "voice"
)

type Preset struct {
	MimeType string
	HW       hwCodec
	Source   string
	Local    string
	Remote   string
}

func NewHardwareCodec(s string) hwCodec {
	switch strings.ToLower(s) {
	case string(Auto):
		return Auto
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
	return fmt.Sprintf("%s %s (%s)", c.MimeType, c.HW, c.Source)
}

func GetPreset(src string, mime string, h hwCodec) (*Preset, error) {
	for _, p := range presets {
		if p.Source == src && p.MimeType == mime && p.HW == h {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("preset %s %s (%s) not found", mime, h, src)
}

func PresetsBySource(src string) []Preset {
	var ps []Preset
	for _, p := range presets {
		if p.Source != src {
			continue
		}
		ps = append(ps, p)
	}
	return ps
}
