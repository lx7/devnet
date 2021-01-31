package signaling

import (
	"testing"
	"time"

	"github.com/lx7/devnet/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSwitch(t *testing.T) {
	receiver := &fakeClient{
		name: "receiver",
		send: make(chan *proto.Frame),
	}

	sw := NewSwitch()
	go sw.Run()
	sw.Register(receiver)

	// define cases
	tests := []struct {
		desc string
		give *proto.Frame
		want *proto.Frame
	}{
		{
			desc: "forward sdp message",
			give: &proto.Frame{
				Src: "sender",
				Dst: "receiver",
			},
			want: &proto.Frame{
				Src: "sender",
				Dst: "receiver",
			},
		},
		{
			desc: "unknown recipient",
			give: &proto.Frame{
				Src: "sender",
				Dst: "unknown recipient",
			},
			want: nil,
		},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// define expectations
			receiver.On("Send").Return()
			time.Sleep(10 * time.Millisecond)

			sw.Forward() <- tt.give
			time.Sleep(10 * time.Millisecond)
			assert.Equal(t, tt.want, receiver.lastmsg)

			receiver.reset()
		})
	}

	sw.Shutdown()
	receiver.AssertExpectations(t)
}

type fakeClient struct {
	mock.Mock
	name    string
	send    chan *proto.Frame
	lastmsg *proto.Frame
}

func (c *fakeClient) Attach(Switch) {
	go func() {
		c.lastmsg = <-c.send
	}()
}

func (c *fakeClient) Send() chan<- *proto.Frame {
	c.Called()
	return c.send
}

func (c *fakeClient) Name() string {
	return c.name
}

func (c *fakeClient) reset() {
	c.lastmsg = nil
}
