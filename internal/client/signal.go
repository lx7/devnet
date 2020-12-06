package client

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lx7/devnet/proto"
	log "github.com/sirupsen/logrus"
)

type SignalSendReceiver interface {
	SignalSender
	SignalReceiver
}

type SignalSender interface {
	Send(proto.Message) error
}

type SignalReceiver interface {
	Receive() <-chan proto.Message
}

// Signal provides signaling via websocket.
type Signal struct {
	conn *websocket.Conn

	recv chan proto.Message
	done chan bool
}

const (
	handshakeTimeout = 20 * time.Second
	connCloseTimeout = 500 * time.Millisecond
	// TODO: enable tls cert verification
	verifyTLS = false
)

func Dial(url string, h http.Header) (*Signal, error) {
	log.Info("dialing: ", url)
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: handshakeTimeout,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: !verifyTLS},
	}

	c, _, err := dialer.Dial(url, h)
	if err != nil {
		return nil, err
	}

	s := &Signal{
		conn: c,
		recv: make(chan proto.Message),
		done: make(chan bool),
	}

	go s.readPump()
	return s, nil
}

func (s *Signal) Send(m proto.Message) error {
	data, err := proto.Marshal(m)
	if err != nil {
		log.Warn("write: ", err)
	}

	err = s.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Signal) Receive() <-chan proto.Message {
	return s.recv
}

func (s *Signal) Close() error {
	data := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	err := s.conn.WriteMessage(websocket.CloseMessage, data)
	if err != nil {
		return err
	}

	select {
	case <-s.done:
	case <-time.After(connCloseTimeout):
		s.conn.Close()
	}
	return nil
}

func (s *Signal) readPump() {
	defer func() {
		close(s.done)
		s.conn.Close()
	}()
	for {
		_, data, err := s.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Error("read message: ", err)
			}
			break
		}

		m, err := proto.Unmarshal(data)
		if err != nil {
			log.Error("unmarshal: ", err)
			continue
		}
		s.recv <- m
	}
	log.Trace("stopping read pump")
}
