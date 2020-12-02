package transport

// MessageReadWriter is the common interface for signaling connections.
// It defines ReadMessage, WriteMessage, Ready and Close.
type MessageReadWriter interface {
	ReadMessage() ([]byte, error)
	WriteMessage([]byte) error
	Close()
}
