package proto

import (
	"encoding/json"
	"fmt"
)

// MessageType defines message types of the JSON signaling protocol
type MessageType string

const (
	MessageTypeSDP  MessageType = "sdp"
	MessageTypeInfo MessageType = "info"
)

// Unmarshal provides unmarshaling for arbitrary JSON message types
func Unmarshal(data []byte) (interface{}, error) {
	peek := struct{ Type MessageType }{}
	if err := json.Unmarshal(data, &peek); err != nil {
		return nil, err
	}

	switch peek.Type {
	case MessageTypeSDP:
		var final SDPMessage
		err := json.Unmarshal(data, &final)
		return final, err
	case MessageTypeInfo:
		var final InfoMessage
		err := json.Unmarshal(data, &final)
		return final, err
	}

	return peek, InvalidMessageTypeError{_type: peek.Type, _json: data}
}

// An InvalidMessageTypeError occurrs when Unmarshal is invoked on a JSON
// literal of unknown type
type InvalidMessageTypeError struct {
	_type MessageType
	_json []byte
}

func (e InvalidMessageTypeError) Error() string {
	return fmt.Sprintf("unknown message type: '%v' json: %s", e._type, e._json)
}

// An UnexpectedMessageTypeError occurrs during Unmarshal when native and JSON
// type can not be mapped
type UnexpectedMessageTypeError struct {
	_type, _exp MessageType
}

func (e UnexpectedMessageTypeError) Error() string {
	return fmt.Sprintf("unexpected message type: '%v' expected: '%s'", e._type, e._exp)
}
