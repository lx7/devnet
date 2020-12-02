package proto

import (
	"testing"

	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestMessage_Marshal(t *testing.T) {
	tests := []struct {
		desc    string
		give    Message
		wantStr string
		wantErr error
	}{
		{
			desc: "marshal SDPMessage",
			give: &SDPMessage{},
			wantStr: `{
				"type":"sdp", 
				"src":"", 
				"dst":"", 
				"sdp":{"sdp":"", "type":"unknown"} 
			}`,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			json, err := Marshal(tt.give)
			assert.IsType(t, tt.wantErr, err)
			assert.JSONEq(t, tt.wantStr, string(json))
		})
	}
}

func TestMessage_Unmarshal(t *testing.T) {
	tests := []struct {
		desc    string
		give    string
		wantMsg Message
		wantErr error
	}{
		{
			desc: "unmarshal SDPMessage",
			give: `{
				"type":"sdp", 
				"src":"", 
				"dst":"", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
			wantMsg: &SDPMessage{
				SDP: webrtc.SessionDescription{
					Type: webrtc.SDPTypeOffer,
					SDP:  "sdp",
				},
			},
			wantErr: nil,
		},
		{
			desc: "unmarshal invalid message",
			give: `{
				"type":"12345", 
				"src":"", 
				"dst":"", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
			wantMsg: nil,
			wantErr: InvalidMessageTypeError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			msg, err := Unmarshal([]byte(tt.give))
			assert.IsType(t, tt.wantErr, err)
			assert.Equal(t, tt.wantMsg, msg)
		})
	}
}
