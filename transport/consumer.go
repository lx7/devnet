package transport

import (
	"github.com/lx7/devnet/proto"
)

// Consumer defines an interface that can be implemented to consume
// messages from the transport layer
type Consumer interface {
	HandleSDP(*proto.SDPMessage)
}
