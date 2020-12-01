package transport

type MessageReadWriter interface {
	Ready() bool
	ReadMessage() (websocket.MessageType, []byte, error)
	WriteMessage(websocket.MessageType, []byte) error
	Close()
}
