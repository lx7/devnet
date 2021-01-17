package client

import (
	"github.com/lx7/devnet/proto"
	"github.com/pion/webrtc/v3"

	"github.com/rs/zerolog/log"
)

// DataChannel provides messaging via webrtc data channel.
type DataChannel struct {
	*webrtc.DataChannel
	remote *webrtc.DataChannel

	conn *webrtc.PeerConnection
	send chan *proto.Frame
	recv chan *proto.Frame
	done chan bool
}

func NewDataChannel(conn *webrtc.PeerConnection) (*DataChannel, error) {
	dc := &DataChannel{
		conn: conn,
		send: make(chan *proto.Frame, 1),
		recv: make(chan *proto.Frame),
		done: make(chan bool),
	}

	c, err := dc.conn.CreateDataChannel("data", nil)
	if err != nil {
		return nil, err
	}
	dc.DataChannel = c
	dc.conn.OnDataChannel(dc.handleRemoteChannel)

	go dc.writePump()
	return dc, nil
}

func (dc *DataChannel) Send(f *proto.Frame) error {
	dc.send <- f
	return nil
}

func (dc *DataChannel) Receive() <-chan *proto.Frame {
	return dc.recv
}

func (dc *DataChannel) Close() error {
	close(dc.done)
	return nil
}

func (dc *DataChannel) writePump() {
	for {
		select {
		case frame := <-dc.send:
			data, err := frame.Marshal()
			if err != nil {
				log.Warn().Err(err).Msg("datachannel: marshal")
				continue
			}

			err = dc.DataChannel.Send(data)
			if err != nil {
				log.Warn().Err(err).Msg("datachannel: write message")
				continue
			}
		case <-dc.done:
			return
		}
	}
}

func (dc *DataChannel) handleRemoteChannel(c *webrtc.DataChannel) {
	log.Info().Uint16("dc_id", *c.ID()).Msg("data channel: new inbound connection")
	dc.remote = c
	dc.remote.OnMessage(dc.handleMessage)
}

func (dc *DataChannel) handleMessage(m webrtc.DataChannelMessage) {
	log.Trace().Interface("msg", m).Msg("datachannel: received message")
	f := &proto.Frame{}
	if err := f.Unmarshal(m.Data); err != nil {
		log.Error().Err(err).Msg("datachannel: unmarshal")
		return
	}
	dc.recv <- f
}
