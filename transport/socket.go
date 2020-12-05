package transport

import (
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Socket implements the MessageReadWriter interface and provides websocket
// connectivity.
type Socket struct {
	sync.RWMutex
	ws           *websocket.Conn
	upgrader     *websocket.Upgrader
	dialer       *websocket.Dialer
	dialerURL    string
	dialerHeader http.Header
	connected    bool
}

const (
	keepaliveInterval    = 10 * time.Second
	reconnectInterval    = 5 * time.Second
	receiveRetryInterval = 100 * time.Millisecond
	sendRetryInterval    = reconnectInterval * 2
	handshakeTimeout     = 20 * time.Second
)

// Upgrade returns a new socket for an incoming http connection.
func Upgrade(w http.ResponseWriter, r *http.Request) (*Socket, error) {
	s := &Socket{
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  0,
			WriteBufferSize: 0,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	log.Trace("upgraded connection from ", ws.RemoteAddr())
	ws.SetCloseHandler(s.closeHandler)
	s.ws = ws
	s.connected = true
	return s, nil
}

// Dial connects to url and returns a new websocket.
func Dial(url string, h http.Header) *Socket {
	s := &Socket{
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: handshakeTimeout,
		},
	}

	// TODO: enable verification as soon as the endpoint has a valid cert
	s.dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	s.dial(url, h)
	return s
}

// ReadMessage reads a single message from the socket
func (s *Socket) ReadMessage() ([]byte, error) {
	if !s.ensureConnected() {
		log.Trace("ReadMessage failed: websocket not connected")
		return nil, ErrWSNotConnected{}
	}

	mt, data, err := s.ws.ReadMessage()
	log.Tracef("read message of type %v from %v", mt, s.ws.RemoteAddr())

	if err != nil && !s.ensureConnected() {
		log.Trace("ReadMessage failed with error: ", err)
		return nil, err
	}
	switch mt {
	case websocket.TextMessage:
		return data, nil
	case -1:
		s.Close()
	default:
		log.Errorf("readmessage: unknown message type %v", mt)
	}
	return nil, nil
}

// WriteMessage writes a single message to the socket. The socket is locked
// during write operations.
func (s *Socket) WriteMessage(data []byte) error {
	if !s.ensureConnected() {
		log.Trace("WriteMessage failed: websocket not connected")
		return ErrWSNotConnected{}
	}
	log.Tracef("write message to %v: %v", s.ws.RemoteAddr(), string(data))
	s.Lock()
	defer s.Unlock()
	return s.ws.WriteMessage(websocket.TextMessage, data)
}

// Close writes a close message to the socket to allow the other endpoint to
// shut down. After a short delay the connection will be terminated.
func (s *Socket) Close() {
	if s.ws != nil && s.Connected() {
		log.Info("closing connection")
		s.Lock()
		if err := s.ws.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		); err != nil {
			log.Error("closing: ", err)
		}

		time.Sleep(10 * time.Millisecond)
		if err := s.ws.Close(); err != nil {
			log.Error("closing: ", err)
		}
		s.Unlock()
	}
	s.setConnected(false)
}

// Connected returns true if a websocket connection is established, else false
func (s *Socket) Connected() bool {
	s.RLock()
	defer s.RUnlock()
	if s.ws == nil {
		s.connected = false
	}

	return s.connected
}

func (s *Socket) dial(url string, h http.Header) {
	log.Info("connecting: ", url)
	s.Lock()
	defer s.Unlock()

	if s.connected {
		return
	}

	s.dialerURL = url
	s.dialerHeader = h

	for {
		ws, _, err := s.dialer.Dial(url, h)
		if err != nil {
			log.Error("dial: ", err)
			time.Sleep(reconnectInterval)
			continue
		}
		ws.SetCloseHandler(s.closeHandler)
		s.ws = ws
		s.connected = true
		s.keepAlive()
		log.Info("connection established")
		time.Sleep(1 * time.Second)
		return
	}
}

func (s *Socket) ensureConnected() bool {
	if !s.Connected() {
		if s.dialer != nil {
			s.Close()
			s.dial(s.dialerURL, s.dialerHeader)
			return s.ensureConnected()
		}
		return false
	}
	return true
}

func (s *Socket) setConnected(state bool) {
	s.Lock()
	defer s.Unlock()

	s.connected = state
}

func (s *Socket) keepAlive() {
	lastResponse := time.Now()
	s.ws.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		for s.Connected() {
			log.Trace("sending ping to ", s.ws.RemoteAddr())
			s.Lock()
			err := s.ws.WriteMessage(websocket.PingMessage, []byte("ping"))
			s.Unlock()
			if err != nil {
				log.Warn("failed to send ping: ", err)
				return
			}
			if time.Since(lastResponse) > keepaliveInterval*2 {
				log.Warn("ping timeout: ", s.ws.RemoteAddr())
				return
			}
			time.Sleep(keepaliveInterval)
		}
	}()
}

func (s *Socket) closeHandler(_ int, _ string) error {
	s.Close()
	return nil
}
