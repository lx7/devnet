package proto

import (
	"testing"

	"github.com/ghodss/yaml"
	"github.com/pion/webrtc/v2"
	"github.com/stretchr/testify/assert"
)

func TestConfig_MarshalYAMLL(t *testing.T) {
	give := &Config{
		Webrtc: &Config_WebRTC{
			Iceservers: []*Config_WebRTC_ICEServer{
				&Config_WebRTC_ICEServer{
					Url: "stun:localhost:19302",
				},
			},
		},
	}

	want := `webrtc:
  iceservers:
  - url: stun:localhost:19302
`

	have, err := yaml.Marshal(give)
	assert.NoError(t, err)
	assert.Equal(t, want, string(have))
}

func TestConfig_UnmarshalYAML(t *testing.T) {
	give := `webrtc:
  iceservers:
  - url: stun:localhost:19302`

	want := &Config{
		Webrtc: &Config_WebRTC{
			Iceservers: []*Config_WebRTC_ICEServer{
				&Config_WebRTC_ICEServer{
					Url: "stun:localhost:19302",
				},
			},
		},
	}

	var have *Config
	err := yaml.Unmarshal([]byte(give), &have)
	assert.NoError(t, err)
	assert.Equal(t, want, have)
}

func TestConfig_ICEServers(t *testing.T) {
	tests := []struct {
		desc string
		give *Config_WebRTC
		want []webrtc.ICEServer
	}{
		{
			desc: "get pion ice servers from wire struct",
			give: &Config_WebRTC{
				Iceservers: []*Config_WebRTC_ICEServer{
					{Url: "turn:devnet.test"},
					{Url: "stun:devnet.test"},
				},
			},
			want: []webrtc.ICEServer{
				{URLs: []string{"turn:devnet.test"}},
				{URLs: []string{"stun:devnet.test"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.EqualValues(t, tt.want, tt.give.ICEServers())
		})
	}
}
