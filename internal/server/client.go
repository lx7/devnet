package server

import (
	"sync"

	"github.com/lx7/devnet/proto"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

// Client defines the client connection handler and attaches to a switch.
type Client interface {
	Attach(Switch)
	Send() chan<- *proto.Frame
	Name() string
}

// DefaultClient implements the Client interface on a websocket connection.
type DefaultClient struct {
	name string
	sw   Switch
	conn *websocket.Conn

	send chan *proto.Frame
}

// NewClient returns a new Client instance.
func NewClient(conn *websocket.Conn, name string) Client {
	c := &DefaultClient{
		name: name,
		conn: conn,

		send: make(chan *proto.Frame, 64),
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
func (c *DefaultClient) Send() chan<- *proto.Frame {
	return c.send
}

func (c *DefaultClient) readPump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseAbnormalClosure,
			) {
				log.Error("read message: ", err)
			}
			break
		}

		f := &proto.Frame{}
		if err := f.Unmarshal(data); err != nil {
			log.Warn("unmarshal message: ", err)
			continue
		}

		c.sw.Forward() <- f
	}
	log.Trace("stopping read pump")
	c.sw.Unregister(c)
}

func (c *DefaultClient) writePump() {
	for f := range c.send {
		data, err := f.Marshal()
		if err != nil {
			log.Warn("marshal message: ", err)
		}

		err = c.conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			log.Warn("write message: ", err)
		}
	}
}
