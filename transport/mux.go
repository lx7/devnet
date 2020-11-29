package transport

import (
	"encoding/json"
	"reflect"

	"github.com/lx7/devnet/proto"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Mux provides parsing and dispatching of signaling messages.
// A consumer implementing the Consumer interface{} can be registered
// to process incoming messages.
type Mux struct {
	socket   *Socket
	consumer Consumer
}

// NewMux returns a new mux instance for s. Consumer interface methods
// are invoked on message retrieval.
func NewMux(s *Socket, c Consumer) *Mux {
	m := &Mux{
		socket:   s,
		consumer: c,
	}
	return m
}

// Send serializes and sends signaling messages through an established socket
// connection.
func (m *Mux) Send(msg proto.Message) error {
	if !m.socket.Connected() {
		return ErrNotConnected
	}

	data, err := json.Marshal(&msg)
	if err != nil {
		return err
	}
	if err = m.socket.WriteMessage(websocket.TextMessage, data); err != nil {
		return err
	}
	return nil
}

// Receive listens on the socket connection and returns on connection close.
// Incoming messages are serialized and dispatched to the consumer.
func (m *Mux) Receive() error {
	if !m.socket.Connected() {
		return ErrNotConnected
	}
	for {
		mt, data, err := m.socket.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return nil
			}
			return err
		}

		switch mt {
		case websocket.TextMessage:
			peek, err := proto.Unmarshal(data)
			if err != nil {
				return err
			}

			switch final := peek.(type) {
			case *proto.SDPMessage:
				m.consumer.HandleSDP(final)
			default:
				log.Warning("unknown native message type: ",
					reflect.TypeOf(final))
			}
		default:
			log.Warning("unknown ws message type: ", mt)
		}
	}
}

// Close terminates the websocket connection and stops the Receive loop.
func (m *Mux) Close() {
	m.socket.Close()
}
