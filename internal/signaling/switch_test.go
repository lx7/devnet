package signaling

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
)

func init() {
	log.SetLevel(log.WarnLevel)
}

func TestSwitch(t *testing.T) {
	// setup
	sender := &mockReadWriter{
		unblock: make(chan bool),
	}
	receiver := &mockReadWriter{}
	defer sender.Close()
	defer receiver.Close()

	sw := NewSwitch()
	go sw.Run()

	// define cases
	tests := []struct {
		desc string
		give []byte
		want []byte
	}{
		{
			desc: "transfer sdp message",
			give: []byte(`{
				"type":"sdp", 
				"src":"sender", 
				"dst":"receiver", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`),
			want: []byte(`{
				"type":"sdp", 
				"src":"sender", 
				"dst":"receiver", 
				"sdp":{ "type":"offer", "sdp":"sdp"} 
			}`),
		},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			// define expectations
			sender.On("ReadMessage").Return(tt.give, nil)
			sender.On("Close").Return()
			receiver.On("WriteMessage", mock.Anything).Return(nil)
			receiver.On("Close").Return()

			// execute
			go sw.Attach(sender, "sender")
			go sw.Attach(receiver, "receiver")

			time.Sleep(10 * time.Millisecond)
			sender.unblock <- true
			time.Sleep(10 * time.Millisecond)

		})
	}

	sw.Shutdown()
	time.Sleep(10 * time.Millisecond)
	sender.AssertExpectations(t)
	receiver.AssertExpectations(t)
}

type mockReadWriter struct {
	mock.Mock
	unblock chan bool
}

func (rw *mockReadWriter) ReadMessage() (m []byte, err error) {
	<-rw.unblock
	res := rw.Called()
	return res.Get(0).([]byte), res.Error(1)
}

func (rw *mockReadWriter) WriteMessage(m []byte) error {
	res := rw.Called(m)
	return res.Error(0)
}

func (rw *mockReadWriter) Close() {
	rw.Called()
}
