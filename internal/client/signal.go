package client

import (
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/lx7/devnet/proto"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

type SignalState int

const (
	SignalStateDisconnected SignalState = iota
	SignalStateConnected
)

func (s SignalState) String() string {
	switch s {
	case SignalStateDisconnected:
		return "disconnected"
	case SignalStateConnected:
		return "connected"
	default:
		return "unknown"
	}
}

type SignalStateHandler func(SignalState)

type SignalSendReceiver interface {
	proto.FrameSender
	proto.FrameReceiver
	HandleStateChange(SignalStateHandler)
}

// Signal provides signaling via websocket.
type Signal struct {
	sync.RWMutex
	url    string
	header http.Header

	conn   *websocket.Conn
	dialer *websocket.Dialer
	send   chan *proto.Frame
	recv   chan *proto.Frame
	done   chan bool
	state  SignalState
	h      SignalStateHandler
}

const (
	handshakeTimeout = 20 * time.Second
	connCloseTimeout = 500 * time.Millisecond

	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	pingInterval      = 15 * time.Second
	pongTimeout       = 20 * time.Second
	reconnectInterval = 10 * time.Second
	maxMessageSize    = 512
	verifyTLS         = true
)

func Dial(url string, h http.Header) *Signal {
	log.Info().Str("url", url).Msg("signaling: dial")
	dialer := &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: handshakeTimeout,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: !verifyTLS},
	}

	s := &Signal{
		url:    url,
		header: h,
		dialer: dialer,
		send:   make(chan *proto.Frame, 1),
		recv:   make(chan *proto.Frame),
		done:   make(chan bool),
	}

	go s.connect()
	time.Sleep(1 * time.Millisecond)

	go s.readPump()
	go s.writePump()
	return s
}

func (s *Signal) connect() {
	s.setState(SignalStateDisconnected)
	s.Lock()
	defer s.Unlock()
	timer := time.NewTimer(0)
	for {
		select {
		case <-timer.C:
			c, _, err := s.dialer.Dial(s.url, s.header)
			if err != nil {
				log.Warn().Err(err).Msg("signaling: dial failed")
				timer.Reset(reconnectInterval)
				continue
			}
			s.conn = c
			s.setState(SignalStateConnected)
			log.Info().Str("url", s.url).Msg("signaling: connected")
			return
		case <-s.done:
			return
		}
	}
}

func (s *Signal) setState(st SignalState) {
	if st != s.state {
		s.state = st
		if s.h != nil {
			s.h(st)
		}
	}
}

func (s *Signal) HandleStateChange(h SignalStateHandler) {
	s.h = h
	h(s.state)
}

func (s *Signal) Send(f *proto.Frame) error {
	s.send <- f
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

func (s *Signal) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case frame := <-s.send:
			data, err := frame.Marshal()
			if err != nil {
				log.Warn().Err(err).Msg("signaling: marshal")
				continue
			}

			s.RLock()
			s.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			err = s.conn.WriteMessage(websocket.TextMessage, data)
			s.RUnlock()
			if err != nil {
				log.Warn().Err(err).Msg("signaling: write message")
				continue
			}

		case <-ticker.C:
			s.RLock()
			s.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			s.RUnlock()
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (s *Signal) readPump() {
	defer func() {
		close(s.done)
		s.setState(SignalStateDisconnected)
	}()
	s.RLock()
	s.conn.SetReadLimit(maxMessageSize)
	s.conn.SetReadDeadline(time.Now().Add(pongTimeout))
	s.RUnlock()
	for {
		s.RLock()
		_, data, err := s.conn.ReadMessage()
		s.RUnlock()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Warn().Err(err).Msg("signaling: read message")
			} else if websocket.IsCloseError(err,
				websocket.CloseNormalClosure,
			) {
				break
			}
			s.connect()
			continue
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
