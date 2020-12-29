package server

import (
	"sync"

	"github.com/lx7/devnet/proto"
	"github.com/spf13/viper"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
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
func NewClient(conn *websocket.Conn, name string) *DefaultClient {
	c := &DefaultClient{
		name: name,
		conn: conn,

		send: make(chan *proto.Frame, 64),
	}
	return c
}

// Configure sends configuration data for ICE servers etc. to the client.
func (c *DefaultClient) Configure(conf *viper.Viper) error {
	var cc *proto.Config
	if err := conf.UnmarshalExact(&cc); err != nil {
		return err
	}

	frame := &proto.Frame{
		Dst:     c.name,
		Payload: &proto.Frame_Config{Config: cc},
	}
	c.Send() <- frame

	return nil
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
				log.Warn().Str("user", c.name).Err(err).Msg("read message")
			}
			break
		}

		f := &proto.Frame{}
		if err := f.Unmarshal(data); err != nil {
			log.Warn().Str("user", c.name).Err(err).Msg("unmarshal message")
			continue
		}

		c.sw.Forward() <- f
	}
	log.Trace().Str("user", c.name).Msg("client read pump done")
	c.sw.Unregister(c)
}

func (c *DefaultClient) writePump() {
	for f := range c.send {
		data, err := f.Marshal()
		if err != nil {
			log.Warn().Str("user", c.name).Err(err).Msg("marshal message")
		}

		err = c.conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			log.Warn().Str("user", c.name).Err(err).Msg("write message")
		}
	}
}
