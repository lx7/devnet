package transport

import (
	"encoding/json"
	"reflect"
	"time"

	"github.com/lx7/devnet/proto"

	log "github.com/sirupsen/logrus"
)

// Mux provides parsing and dispatching of signaling messages.
// A consumer implementing the signaler interface{} can be registered
// to process incoming messages.
type Mux struct {
	ReadWriter MessageReadWriter
	Consumer   Consumer
}

// Send serializes and sends signaling messages through the MessageReadWriter
// interface
func (m *Mux) Send(msg proto.Message) error {
	data, err := json.Marshal(&msg)
	if err != nil {
		return err
	}
	if err = m.ReadWriter.WriteMessage(data); err != nil {
		return err
	}
	return nil
}

func (m *Mux) SendRetry(msg proto.Message, max int, wait time.Duration) error {
	var err error
	for n := 1; n <= max; n++ {
		if err = m.Send(msg); err != nil {
			time.Sleep(time.Duration(n) * wait)
		}
	}
	if err != nil {
		return ErrRetryCountExceeded{retries: max, lasterr: err}
	}
	return nil
}

// Receive reads messages from the MessageReadWriter instance and returns on
// close. Incoming messages are serialized and dispatched to the Signaler.
func (m *Mux) Receive() error {
	defer func() {
		m.Consumer.HandleClose()
		m.ReadWriter.Close()
	}()
	for {
		data, err := m.ReadWriter.ReadMessage()
		if err != nil {
			return err
		}

		if data != nil {
			m.dispatch(data)
		}
	}
}

func (m *Mux) dispatch(data []byte) {
	peek, err := proto.Unmarshal(data)
	if err != nil {
		log.Warnf("invalid message: %s", data)
		return
	}

	switch final := peek.(type) {
	case *proto.SDPMessage:
		m.Consumer.HandleSDP(final)
	default:
		log.Warning("unknown message type: ", reflect.TypeOf(final))
	}
}
