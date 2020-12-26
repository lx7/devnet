package gst

import (
	"fmt"
)

type codecKind string

const (
	Audio codecKind = "audio"
	Video codecKind = "video"
)

type codecName string

const (
	H264 codecName = "H264"
	Opus codecName = "Opus"
)

type codecAccel string

const (
	Software codecAccel = "Software"
	VAAPI    codecAccel = "VAAPI"
	NVCODEC  codecAccel = "NVCODEC"
	VDPAU    codecAccel = "VDPAU"
	OSXVT    codecAccel = "OSXVT"
)

type sourceType string

const (
	Screen sourceType = "screen"
	Camera sourceType = "camera"
	Voice  sourceType = "voice"
)

const (
	ClockRateVideo float32 = 90000
	ClockRateAudio float32 = 48000
)

type Preset struct {
	Kind        codecKind
	Codec       codecName
	Accel       codecAccel
	Source      sourceType
	Local       string
	Remote      string
	Clock       float32
	PayloadType uint8
}

func (c *Preset) String() string {
	return fmt.Sprintf("%s/%s (%s)", c.Codec, c.Accel, c.Source)
}

func GetPreset(s sourceType, c codecName, a codecAccel) (*Preset, error) {
	for _, p := range presets {
		if p.Source == s && p.Codec == c && p.Accel == a {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("preset %s/%s (%s) not found", c, a, s)
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
