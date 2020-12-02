package transport

import (
	"testing"
	"time"

	"github.com/lx7/devnet/proto"

	"github.com/pion/webrtc/v2"
	"github.com/stretchr/testify/mock"
)

func TestMux(t *testing.T) {
	// set up mocks
	rw := &mockReadWriter{
		echo: make(chan []byte),
	}
	defer rw.Close()

	consumer := &mockConsumer{}

	// create new mux instance and start main loop
	mux := &Mux{
		ReadWriter: rw,
		Consumer:   consumer,
	}

	go func() {
		err := mux.Receive()
		if err != nil {
			t.Error(err)
		}
	}()

	msg := &proto.SDPMessage{
		Src: "",
		Dst: "",
		SDP: webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  "sdp",
		},
	}

	// define expectations
	consumer.On("HandleSDP", msg).Return()

	// run test
	mux.Send(msg)
	time.Sleep(10 * time.Millisecond)

	// verify the expectations were met
	consumer.AssertExpectations(t)
}

type mockReadWriter struct {
	mock.Mock
	echo chan []byte
}

func (s *mockReadWriter) Ready() bool {
	return true
}

func (s *mockReadWriter) ReadMessage() (m []byte, err error) {
	m = <-s.echo
	return m, nil
}

func (s *mockReadWriter) WriteMessage(m []byte) error {
	s.echo <- m
	return nil
}

func (s *mockReadWriter) Close() {
}

type mockConsumer struct {
	mock.Mock
}

func (s *mockConsumer) HandleSDP(m *proto.SDPMessage) {
	s.Called(m)
}

func (s *mockConsumer) HandleClose() {
}
