package proto

import (
	"testing"

	"github.com/pion/webrtc/v2"
	"github.com/stretchr/testify/assert"
)

func TestSDP_ToPion(t *testing.T) {
	tests := []struct {
		desc string
		give *Frame_Sdp
		want webrtc.SessionDescription
	}{
		{
			desc: "convert wire SDP message to Pion",
			give: &Frame_Sdp{&SDP{
				Type: SDP_OFFER,
				Desc: "sdp",
			}},
			want: webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  "sdp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.want, ToPion(tt.give))
		})
	}
}

func TestSDP_PayloadWithPion(t *testing.T) {
	tests := []struct {
		desc string
		give webrtc.SessionDescription
		want *Frame_Sdp
	}{
		{
			desc: "convert wire SDP message to Pion",
			give: webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  "sdp",
			},
			want: &Frame_Sdp{&SDP{
				Type: SDP_OFFER,
				Desc: "sdp",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			assert.Equal(t, tt.want, WithPion(tt.give))
		})
	}
}
