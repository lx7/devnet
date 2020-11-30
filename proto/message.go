package proto

import (
	"encoding/json"
)

type Message interface {
	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
}

// Unmarshal provides unmarshaling for arbitrary JSON message types
func Unmarshal(data []byte) (Message, error) {
	peek := struct{ Type MessageType }{}
	if err := json.Unmarshal(data, &peek); err != nil {
		return nil, err
	}

	switch peek.Type {
	case MessageTypeSDP:
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
