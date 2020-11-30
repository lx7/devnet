package proto

import (
	"encoding/json"

	"github.com/pion/webrtc/v2"
)

// SDPMessage implements a routable container for Session Description Protocol
// messages on the WebSocket transport layer
type SDPMessage struct {
	Src string                    `json:"src"`
	Dst string                    `json:"dst"`
	SDP webrtc.SessionDescription `json:"sdp"`
}

// MarshalJSON provides JSON marshaling for SDPMessage
func (m *SDPMessage) MarshalJSON() ([]byte, error) {
	type Alias SDPMessage

	return json.Marshal(&struct {
		*Alias
		Type MessageType `json:"type"`
	}{
		Alias: (*Alias)(m),
		Type:  MessageTypeSDP,
	})
}

// UnmarshalJSON provides JSON unmarshaling for SDPMessage
func (m *SDPMessage) UnmarshalJSON(data []byte) error {
	type Alias SDPMessage
	aux := &struct {
		*Alias
		Type MessageType `json:"type"`
	}{}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Type != MessageTypeSDP {
		return UnexpectedMessageTypeError{_type: aux.Type, _exp: MessageTypeSDP}
	}
	*m = (SDPMessage)(*aux.Alias)
	return nil
}
