package client

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/lx7/devnet/proto"
)

type Event interface{}

type EventConnected struct{}

type EventDisconnected struct{}

type EventSessionStart struct{}

type EventSessionEnd struct{}

type EventSCInboundStart struct{}

type EventSCInboundEnd struct{}

type EventCameraInboundStart struct{}

type EventCameraInboundEnd struct{}

type EventRCon struct {
	Data *proto.Control
}

func (s *Session) Handle(f interface{}, args ...interface{}) error {
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

func (s *Session) callHandler(e Event) bool {
	h, ok := s.h[reflect.TypeOf(e)]
	if ok {
		vargs := append([]reflect.Value{reflect.ValueOf(e)}, h.args...)
		h.f.Call(vargs)
		return true
	}
	return false
}
