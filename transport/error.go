package transport

import (
	"fmt"
	"net"
)

// An ErrRetryCountExceeded occurrs when the requested action was excecuted
// for the requested number of times without success.
type ErrRetryCountExceeded struct {
	retries int
	lasterr error
}

func (e ErrRetryCountExceeded) Error() string {
	return fmt.Sprintf("failed %v times, last error: %v", e.retries, e.lasterr)
}

// ErrWSNotConnected occurrs on invocation of ReadMessage() or
// WriteMessage() without the WebSocket connection already established.
type ErrWSNotConnected struct{}

func (e ErrWSNotConnected) Error() string {
	return "websocket not connected"
}

//ErrInvalidWSMessagetype is caused by a websocket message of unknown type.
type ErrInvalidWSMessageType struct {
	mt  int
	src net.Addr
}

func (e ErrInvalidWSMessageType) Error() string {
	return fmt.Sprintf("invalid ws message type %v from %s", e.mt, e.src)
}
