package proto

import (
	"encoding/json"
)

// Message defines the common interface for messages that can be sent over
// the signaling channel.
type Message interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
}

// ---------------------------------------------------------------------------
// interface
// ---------------------------------------------------------------------------

// Unmarshal provides unmarshaling for arbitrary JSON message types
func Unmarshal(data []byte) (Message, error) {
	peek := struct{ Type messageType }{}
	if err := json.Unmarshal(data, &peek); err != nil {
		return nil, err
	}

	switch peek.Type {
	case messageTypeSDP:
		final := &SDPMessage{}
		err := json.Unmarshal(data, &final)
		return final, err
	}

	return nil, InvalidMessageTypeError{_type: peek.Type, _json: data}
}

// Marshal provides marshaling of arbitrary JSON message types
func Marshal(m Message) ([]byte, error) {
	return m.MarshalJSON()
}
