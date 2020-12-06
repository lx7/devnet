package testutil

import (
	"io/ioutil"
	"testing"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(ioutil.Discard)
}

func TestGlobalHook(t *testing.T) {
	hook := NewLogHook()

	if got := hook.Entry(log.ErrorLevel); got != nil {
		t.Errorf("no logs in history: exp: %v got: %v", nil, got)
	}

	log.Error("error")

	exp := log.ErrorLevel
	got := hook.Entry(log.ErrorLevel)
	if got == nil {
		t.Error("error logged: expected log entry, got nil")
	} else if got := hook.Entry(log.ErrorLevel); got.Level != exp {
		t.Errorf("error logged: exp: %v got: %v", exp, got.Level)
	}

	hook.Reset()

	if got := hook.Entry(log.ErrorLevel); got != nil {
		t.Errorf("no logs in history: exp: %v got: %v", nil, got)
	}
}
