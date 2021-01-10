package proto

import (
	"testing"

	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
)

func TestICE_ICECandidate(t *testing.T) {
	tests := []struct {
		desc string
		give *Frame_Ice
		want webrtc.ICECandidateInit
	}{
		{
			desc: "convert wire ICE message to Pion",
			give: &Frame_Ice{&ICE{
				Candidate:        "candidate:123",
				SdpMid:           "0",
				SdpMLineIndex:    uint32(0),
				UsernameFragment: "def",
			}},
			want: webrtc.ICECandidateInit{
				Candidate:        "candidate:123",
				SDPMid:           refString("0"),
				SDPMLineIndex:    refUint16(0),
				UsernameFragment: refString("def"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.give.ICECandidate())
		})
	}
}

func TestSDP_PayloadWithICECandidate(t *testing.T) {
	tests := []struct {
		desc string
		give webrtc.ICECandidateInit
		want *Frame_Ice
	}{
		{
			give: webrtc.ICECandidateInit{
				Candidate:        "candidate:123",
				SDPMid:           refString("0"),
				SDPMLineIndex:    refUint16(0),
				UsernameFragment: refString("def"),
			},
			want: &Frame_Ice{&ICE{
				Candidate:        "candidate:123",
				SdpMid:           "0",
				SdpMLineIndex:    uint32(0),
				UsernameFragment: "def",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.want, PayloadWithICECandidate(tt.give))
		})
	}
}

func refString(s string) *string {
	return &s
}

func refUint16(i uint16) *uint16 {
	return &i
}
