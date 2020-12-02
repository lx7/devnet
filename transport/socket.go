package transport

import (
	"crypto/tls"
	"errors"
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
	keepaliveInterval    = 25 * time.Second
	reconnectInterval    = 5 * time.Second
	receiveRetryInterval = 100 * time.Millisecond
	sendRetryInterval    = reconnectInterval * 2
	handshakeTimeout     = 20 * time.Second
)

var (
	// ErrWSNotConnected occurrs on invocation of ReadMessage() or
	// WriteMessage() without the WebSocket connection already established.
	ErrWSNotConnected = errors.New("websocket: not connected")

	//ErrInvalidWSMessagetype is caused by a websocket message of unknown type.
	ErrInvalidWSMessageType = errors.New("websocket: invalid message type")
)

// ---------------------------------------------------------------------------
// constructors
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// interface
// ---------------------------------------------------------------------------

// ReadMessage reads a single message from the socket
func (s *Socket) ReadMessage() (data []byte, err error) {
	if !s.Connected() {
		return data, ErrWSNotConnected
	}
	mt, data, err := s.ws.ReadMessage()

	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Info("connection closed")
		}
		if s.dialer != nil {
			s.reconnect()
		}
		return nil, nil
	}

	if mt != websocket.TextMessage {
		return nil, ErrInvalidWSMessageType
	}

	return data, err
}

// WriteMessage writes a single message to the socket. The socket is locked
// during write operations.
func (s *Socket) WriteMessage(data []byte) error {
	log.Trace("send data: ", string(data))
	if !s.Connected() {
		return ErrWSNotConnected
	}
	s.Lock()
	err := s.ws.WriteMessage(websocket.TextMessage, data)
	s.Unlock()
	if err != nil && s.dialer != nil {
		s.reconnect()
	}
	return err
}

// Close writes a close message to the socket to allow the other endpoint to
// shut down. After a short delay the connection will be terminated.
func (s *Socket) Close() {
	s.Lock()
	if s.ws != nil {
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
	}
	s.Unlock()

	s.setConnected(false)
}

// ---------------------------------------------------------------------------
// exported
// ---------------------------------------------------------------------------

// Connected returns true if a websocket connection is established, else false
func (s *Socket) Connected() bool {
	s.RLock()
	defer s.RUnlock()

	return s.connected
}

// ---------------------------------------------------------------------------
// utility
// ---------------------------------------------------------------------------

func (s *Socket) dial(url string, h http.Header) {
	log.Info("connecting: ", url)
	s.dialerURL = url
	s.dialerHeader = h

	go func() {
		for {
			ws, _, err := s.dialer.Dial(url, h)
			if err == nil {
				ws.SetCloseHandler(s.closeHandler)
				s.Lock()
				s.ws = ws
				s.connected = true
				s.keepAlive()
				s.Unlock()
				return
			}
			log.Error("dial: ", err)
			time.Sleep(reconnectInterval)
		}
	}()
}

func (s *Socket) reconnect() {
	s.Close()
	s.dial(s.dialerURL, s.dialerHeader)
}

func (s *Socket) keepAlive() {
	lastResponse := time.Now()
	s.ws.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		for {
			s.Lock()
			err := s.ws.WriteMessage(websocket.PingMessage, []byte("ping"))
			s.Unlock()
			if err != nil {
				s.Close()
				return
			}
			if time.Since(lastResponse) > keepaliveInterval*2 {
				s.Close()
				return
			}
			time.Sleep(keepaliveInterval)
		}
	}()
}

func (s *Socket) setConnected(state bool) {
	s.Lock()
	defer s.Unlock()

	s.connected = state
}

func (s *Socket) closeHandler(_ int, _ string) error {
	s.setConnected(false)
	return nil
}
