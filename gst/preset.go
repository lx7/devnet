package gst

import (
	"fmt"
)

type codecKind string

const (
	AUDIO codecKind = "audio"
	VIDEO codecKind = "video"
)

type codecName string

const (
	H264 codecName = "H264"
	Opus codecName = "Opus"
)

type codecAccel string

const (
	AccelTypeNone  codecAccel = "SW"
	AccelTypeVAAPI codecAccel = "VAAPI"
	AcceltypeVDPAU codecAccel = "VDPAU"
	AcceltypeOSXVT codecAccel = "VideoToolbox"
)

type sourceType string

const (
	SourceTypeScreen sourceType = "screen"
	SourceTypeCamera sourceType = "camera"
	SourceTypeVoice  sourceType = "voice"
)

const (
	ClockRateVideo float32 = 90000
	ClockRateAudio float32 = 48000
)

type Preset struct {
	CodecKind   codecKind
	CodecName   codecName
	Accel       codecAccel
	SourceType  sourceType
	Local       string
	Remote      string
	Clock       float32
	PayloadType uint8
}

func (c *Preset) String() string {
	return fmt.Sprintf("%s/%s (%s)", c.CodecName, c.Accel, c.SourceType)
}

func PresetBySource(t sourceType, n codecName, a codecAccel) (*Preset, error) {
	for _, p := range presets {
		if p.SourceType == t && p.CodecName == n && p.Accel == a {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("preset %s/%s (%s) not found", n, a, t)
}

func PresetsBySource(t sourceType) []Preset {
	var ps []Preset
	for _, p := range presets {
		if p.SourceType != t {
			continue
		}
		ps = append(ps, p)
	}
	return ps
}

type presetMap map[codecName]map[codecAccel]*Preset

var vPresets, sPresets, cPresets presetMap

func mkPresetMap(t sourceType) presetMap {
	m := make(presetMap)
	for _, p := range presets {
		if p.SourceType != t {
			continue
		}
		if m[p.CodecName] == nil {
			m[p.CodecName] = make(map[codecAccel]*Preset)
		}
		m[p.CodecName][p.Accel] = &p
	}
	return m
}

func init() {
	vPresets = mkPresetMap(SourceTypeVoice)
	sPresets = mkPresetMap(SourceTypeScreen)
	cPresets = mkPresetMap(SourceTypeCamera)
}
