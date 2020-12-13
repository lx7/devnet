package proto

import (
	pb "github.com/golang/protobuf/proto"
)

func (f *Frame) Marshal() ([]byte, error) {
	return pb.Marshal(f)
}

func (f *Frame) Unmarshal(in []byte) error {
	return pb.Unmarshal(in, f)
}
