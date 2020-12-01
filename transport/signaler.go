package client

import "github.com/lx7/devnet/proto"

// Signaler defines an interface that can be implemented to transfer messages
// through the transport layer
type Signaler interface {
	Send(dst string, m *proto.Message)
	HandleSDP(m *proto.SDPMessage)
}
