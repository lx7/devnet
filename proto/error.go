package proto

import (
	"fmt"
)

// An InvalidMessageTypeError occurrs when Unmarshal is invoked on a JSON
// literal of unknown type
type InvalidMessageTypeError struct {
	_type MessageType
	_json []byte
}

func (e InvalidMessageTypeError) Error() string {
	return fmt.Sprintf("unknown message type: '%v' json: '%s'", e._type, e._json)
}

// An UnexpectedMessageTypeError occurrs during Unmarshal when native and JSON
// type can not be mapped
type UnexpectedMessageTypeError struct {
	_type, _exp MessageType
}

func (e UnexpectedMessageTypeError) Error() string {
	return fmt.Sprintf("unexpected message type: '%v' expected: '%s'", e._type, e._exp)
}
