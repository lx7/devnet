package proto

import (
	"encoding/json"
	"testing"

	"github.com/pion/webrtc/v2"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestSDPMessage_Marshal(t *testing.T) {
	tests := []struct {
		desc    string
		give    SDPMessage
		wantStr string
		wantErr error
	}{
		{
			desc: "marshal SDPMessage",
			give: SDPMessage{
				Src: "user 1",
				Dst: "user 2",
				SDP: webrtc.SessionDescription{
					Type: webrtc.SDPTypeOffer,
					SDP:  "sdp",
				},
			},
			wantStr: `{
				"type":"sdp", 
				"src":"user 1", 
				"dst":"user 2", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			json, err := tt.give.MarshalJSON()
			assert.IsType(t, tt.wantErr, err)
			assert.JSONEq(t, tt.wantStr, string(json))
		})
	}
}

func TestSDPMessage_Unmarshal(t *testing.T) {
	tests := []struct {
		desc    string
		give    string
		wantMsg SDPMessage
		wantErr error
	}{
		{
			desc: "unmarshal SDPMessage",
			give: `{
				"type":"sdp", 
				"src":"user 1", 
				"dst":"user 2", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
			wantMsg: SDPMessage{
				Src: "user 1",
				Dst: "user 2",
				SDP: webrtc.SessionDescription{
					Type: webrtc.SDPTypeOffer,
					SDP:  "sdp",
				},
			},
			wantErr: nil,
		},
		{
			desc: "unmarshal invalid type",
			give: `{
				"type":"12345", 
				"src":"user 1", 
				"dst":"user 2", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`,
			wantMsg: SDPMessage{},
			wantErr: UnexpectedMessageTypeError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var msg SDPMessage
			err := json.Unmarshal([]byte(tt.give), &msg)
			assert.IsType(t, tt.wantErr, err)
			assert.Equal(t, tt.wantMsg, msg)
		})
	}
}
