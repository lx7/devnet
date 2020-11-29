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

const (
	KeepaliveInterval    = 25 * time.Second
	ReconnectInterval    = 5 * time.Second
	ReceiveRetryInterval = 100 * time.Millisecond
	SendRetryInterval    = ReconnectInterval * 2
	HandshakeTimeout     = 20 * time.Second
)

var (
	ErrNotConnected = errors.New("websocket: not connected")
)

// Upgrade returns a new websocket for an incoming http connection.
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
	err := s.Upgrade(w, r, nil)
	return s, err
}

// Dial connects to url and returns a new websocket.
func Dial(url string, h http.Header) *Socket {
	s := &Socket{
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: HandshakeTimeout,
		},
	}

	// TODO: enable verification as soon as the endpoint has a valid cert
	s.dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	s.Dial(url, h)
	return s
}

// Socket provides websocket connectivity.
type Socket struct {
	ws           *websocket.Conn
	upgrader     *websocket.Upgrader
	dialer       *websocket.Dialer
	dialerUrl    string
	dialerHeader http.Header
	mu           sync.RWMutex
	connected    bool
}

// Upgrade upgrades an existing http connection to the websocket protocol.
// The socket is locked during upgrade.
func (s *Socket) Upgrade(w http.ResponseWriter, r *http.Request, h http.Header) error {
	ws, err := s.upgrader.Upgrade(w, r, h)
	if err != nil {
		log.Error("upgrade: ", err)
		return err
	}
	s.mu.Lock()
	s.ws = ws
	s.connected = true
	s.mu.Unlock()
	return nil
}

// Dial requests url to establish a websocket connection. The socket is locked
// while the connection is being established.
func (s *Socket) Dial(url string, h http.Header) {
	log.Info("connecting: ", url)
	s.dialerUrl = url
	s.dialerHeader = h

	go func() {
		for {
			ws, _, err := s.dialer.Dial(url, h)
			if err == nil {
				s.mu.Lock()
				s.ws = ws
				s.connected = true
				s.keepAlive()
				s.mu.Unlock()
				return
			} else {
				log.Error("websocket: ", err)
				time.Sleep(ReconnectInterval)
			}
		}
	}()
}

// CloseAndDial closes s and requests url to establish a new websocket
// connection.
func (s *Socket) CloseAndDial() {
	s.Close()
	s.Dial(s.dialerUrl, s.dialerHeader)
}

// Close writes a close message to the socket to allow the other endpoint to
// shut down. After a short delay the connection will be terminated.
func (s *Socket) Close() {
	s.mu.Lock()
	if s.ws != nil {
		s.ws.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		select {
		case <-time.After(time.Second):
		}
		s.ws.Close()
	}
	s.mu.Unlock()

	s.setConnected(false)
}

// Connected returns true if a connection is established, otherwise false.
func (s *Socket) Connected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.connected
}

// ReadMessage reads a single message from the socket
func (s *Socket) ReadMessage() (mt int, data []byte, err error) {
	err = ErrNotConnected
	if s.Connected() {
		mt, data, err = s.ws.ReadMessage()
		if err == nil {
			log.Trace("recv data: ", string(data))
		} else if s.dialer != nil {
			s.CloseAndDial()
		}
	}
	return
}

// WriteMessage writes a single message to the socket. The socket is locked
// during write operations.
func (s *Socket) WriteMessage(mt int, data []byte) error {
	log.Trace("send data: ", string(data))
	if s.Connected() {
		s.mu.Lock()
		err := s.ws.WriteMessage(mt, data)
		s.mu.Unlock()
		if err != nil && s.dialer != nil {
			s.CloseAndDial()
		}
		return err
	}
	return ErrNotConnected
}

func (s *Socket) keepAlive() {
	lastResponse := time.Now()
	s.ws.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		for {
			s.mu.Lock()
			err := s.ws.WriteMessage(websocket.PingMessage, []byte("keepalive"))
			s.mu.Unlock()
			if err != nil {
				s.Close()
				return
			}
			if time.Now().Sub(lastResponse) > KeepaliveInterval*2 {
				s.Close()
				return
			}
			time.Sleep(KeepaliveInterval)
		}
	}()
}

func (s *Socket) setConnected(state bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connected = state
}
