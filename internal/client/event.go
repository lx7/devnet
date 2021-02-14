package client

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/lx7/devnet/proto"
)

type Event interface{}

// EventConnected occurs when the connection to the signaling service has been
// established.
type EventConnected struct{}

// EventDisconnected occurs when the connection to the signaling service has
// been closed.
type EventDisconnected struct{}

// EventPeerConnected occurs when a new peer connection has been established.
type EventPeerConnected struct {
	Peer Peer
}

// EventPeerDisconnected occurs when a peer has been disconnected.
type EventPeerDisconnected struct {
	Peer Peer
}

// EventPeerDisconnected occurs when a peer session has been closed.
type EventPeerClosed struct {
	Peer Peer
}

type EventStreamStart struct {
	Peer   Peer
	Stream Stream
}

type EventStreamEnd struct {
	Peer   Peer
	Stream Stream
}

type EventRCon struct {
	Peer Peer
	Data *proto.Control
}

func (s *DefaultSession) Handle(f interface{}, args ...interface{}) error {
	rf := reflect.ValueOf(f)
	kind := rf.Type().Kind()
	if kind != reflect.Func {
		return fmt.Errorf("type mismatch. want: func, have: %v", kind)
	}

	rt := rf.Type()
	if rt.NumIn()-1 != len(args) {
		return errors.New("arity mismatch")
	}
	for i := range args {
		if rt.In(i+1) != reflect.TypeOf(args[i]) {
			return fmt.Errorf("arg #%v type mismatch", i)
		}
	}

	vargs := make([]reflect.Value, len(args))
	for i := range args {
		vargs[i] = reflect.ValueOf(args[i])
	}

	s.h[rt.In(0)] = handler{rf, vargs}
	return nil
}

type handler struct {
	f    reflect.Value
	args []reflect.Value
}

func (s *DefaultSession) callHandler(e Event) bool {
	h, ok := s.h[reflect.TypeOf(e)]
	if ok {
		vargs := append([]reflect.Value{reflect.ValueOf(e)}, h.args...)
		h.f.Call(vargs)
		return true
	}
	return false
}
