package proto

import (
	pb "google.golang.org/protobuf/proto"
)

func (f *Frame) Marshal() ([]byte, error) {
	return pb.Marshal(f)
}

func (f *Frame) Unmarshal(in []byte) error {
	return pb.Unmarshal(in, f)
}

type FrameSendReceiver interface {
	FrameSender
	FrameReceiver
}

type FrameSender interface {
	Send(*Frame) error
}

type FrameReceiver interface {
	Receive() <-chan *Frame
}
