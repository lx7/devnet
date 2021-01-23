package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEvent(t *testing.T) {
	s, err := NewSession("testuser", &fakeSignal{})
	assert.NoError(t, err)

	c := &fakeEventConsumer{}

	// define expectations
	c.On("handleEvent", EventConnected{}, 10).Return()

	err = s.Handle(10)
	assert.Error(t, err, "wrong type should cause an error")

	err = s.Handle(c.handleEvent, 10, 100)
	assert.Error(t, err, "wrong number of args should cause an error")

	err = s.Handle(c.handleEvent, "string")
	assert.Error(t, err, "wrong arg type should cause an error")

	err = s.Handle(c.handleEvent, 10)
	assert.NoError(t, err)

	ok := s.callHandler(EventDisconnected{})
	assert.False(t, ok, "unregistered event should return false")

	ok = s.callHandler(EventConnected{})
	assert.True(t, ok, "registered event should return true")

	c.AssertExpectations(t)
}

type fakeEventConsumer struct {
	mock.Mock
}

func (c *fakeEventConsumer) handleEvent(e EventConnected, i int) {
	c.Called(e, i)
}
