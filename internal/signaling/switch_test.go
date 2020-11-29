package signaling

import (
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.ErrorLevel)
}

func TestSwitch(t *testing.T) {
	sw := NewSwitch()
	go sw.Run()

	c1 := sw.Attach(nil, "client1")
	c2 := NewClient(nil, sw, "client2")
	sw.Register(c2)

	time.Sleep(100 * time.Millisecond)

	sw.Unregister(c1)
	sw.Unregister(c2)
	sw.Shutdown()

	// TODO: improve test coverage
}
