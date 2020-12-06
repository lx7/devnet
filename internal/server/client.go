package server

import (
	"reflect"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/lx7/devnet/proto"

	log "github.com/sirupsen/logrus"
)

// Client defines the client connection handler and attaches to a switch.
type Client interface {
	Attach(Switch)
	Send() chan<- *proto.SDPMessage
	Name() string
}

// DefaultClient implements the Client interface on a websocket connection.
type DefaultClient struct {
	name string
	sw   Switch
	conn *websocket.Conn

	send chan *proto.SDPMessage
}

// NewClient returns a new Client instance.
func NewClient(conn *websocket.Conn, name string) Client {
	c := &DefaultClient{
		name: name,
		conn: conn,

		send: make(chan *proto.SDPMessage, 64),
	}
	return c
}

// Attach connects to a switch and starts message processing. Returns on
// connection close.
func (c *DefaultClient) Attach(sw Switch) {
	c.sw = sw

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		c.readPump()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		c.writePump()
		wg.Done()
	}()
	wg.Wait()
}

// Name returns the client name.
func (c *DefaultClient) Name() string {
	return c.name
}

// Send sends a message through the network connection.
func (c *DefaultClient) Send() chan<- *proto.SDPMessage {
	return c.send
}

func (c *DefaultClient) readPump() {
	defer func() {
		close(c.send)
		c.conn.Close()
	}()
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if unexpectedCloseError(err) {
				log.Error("read message: ", err)
			}
			break
		}

		m, err := proto.Unmarshal(data)
		if err != nil {
			log.Error("unmarshal: ", err)
			continue
		}

		switch final := m.(type) {
		case *proto.SDPMessage:
			c.sw.Forward() <- final
		default:
			log.Warn("unexpected message type: ", reflect.TypeOf(final))
		}
	}
	log.Trace("stopping read pump")
}

func (c *DefaultClient) writePump() {
	for m := range c.send {
		data, err := proto.Marshal(m)
		if err != nil {
			log.Warn("write: ", err)
		}

		err = c.conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Error("write: ", err)
		}
	}
}

func unexpectedCloseError(err error) bool {
	return websocket.IsUnexpectedCloseError(
		err,
		websocket.CloseGoingAway,
		websocket.CloseAbnormalClosure,
	)
}
