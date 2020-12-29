package client

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/lx7/devnet/proto"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type SignalSendReceiver interface {
	SignalSender
	SignalReceiver
}

type SignalSender interface {
	Send(*proto.Frame) error
}

type SignalReceiver interface {
	Receive() <-chan *proto.Frame
}

// Signal provides signaling via websocket.
type Signal struct {
	conn *websocket.Conn

	recv chan *proto.Frame
	done chan bool
}

const (
	handshakeTimeout = 20 * time.Second
	connCloseTimeout = 500 * time.Millisecond
	// TODO: enable tls cert verification
	verifyTLS = false
)

func Dial(url string, h http.Header) (*Signal, error) {
	log.Info().Str("url", url).Msg("signaling: dial")
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: handshakeTimeout,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: !verifyTLS},
	}

	c, _, err := dialer.Dial(url, h)
	if err != nil {
		return nil, err
	}
	log.Info().Str("url", url).Msg("signaling: connected")

	s := &Signal{
		conn: c,
		recv: make(chan *proto.Frame),
		done: make(chan bool),
	}

	go s.readPump()
	return s, nil
}

func (s *Signal) Send(f *proto.Frame) error {
	data, err := f.Marshal()
	if err != nil {
		log.Warn().Err(err).Msg("signaling: write")
	}

	err = s.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *Signal) Receive() <-chan *proto.Frame {
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
				log.Error().Err(err).Msg("signaling: read message")
			}
			break
		}

		f := &proto.Frame{}
		if err := f.Unmarshal(data); err != nil {
			log.Error().Err(err).Msg("signaling: unmarshal")
			continue
		}
		s.recv <- f
	}
	log.Info().Msg("signaling: done")
}
