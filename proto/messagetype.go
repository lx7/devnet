package proto

// MessageType defines message types of the JSON signaling protocol
type MessageType string

const (
	MessageTypeSDP  MessageType = "sdp"
	MessageTypeInfo MessageType = "info"
)
