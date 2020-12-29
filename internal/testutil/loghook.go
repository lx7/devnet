package testutil

import (
	"sync"

	"github.com/rs/zerolog"
)

// LogEntry represents an entry in the log history
type LogEntry struct {
	Level zerolog.Level
	Msg   string
}

// LogHook implements the zerolog hook interface to provides data on logged
// errors for testing.
type LogHook struct {
	sync.RWMutex
	Entries []LogEntry
}

// Entry checks the recorded log entries for severity. Returns the
// last corrensponding entry if level >= minlevel.
func (h *LogHook) Entry(minlevel zerolog.Level) *LogEntry {
	h.RLock()
	defer h.RUnlock()
	for _, e := range h.Entries {
		if e.Level >= minlevel {
			return &e
		}
	}
	return nil
}

// Run implements the zerolog.Hook interface
func (h *LogHook) Run(e *zerolog.Event, l zerolog.Level, msg string) {
	h.Lock()
	defer h.Unlock()
	h.Entries = append(h.Entries, LogEntry{
		Level: l,
		Msg:   msg,
	})
}

// Reset clears the history of log entries in this LogHook instance.
func (h *LogHook) Reset() {
	h.Lock()
	defer h.Unlock()
	h.Entries = make([]LogEntry, 0)
}
