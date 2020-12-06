package testutil

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// LogHook implements the logrus hook interface to provide data on logged
// errors for testing.
type LogHook struct {
	Entries []log.Entry
	mu      sync.RWMutex
}

// NewLogHook creates a new LogHook instance and adds it to the global logger.
func NewLogHook() *LogHook {
	h := new(LogHook)
	log.AddHook(h)
	return h
}

// Entry checks all recorded log entries for severity. Returns the
// corrensponding entry if level <= maxlevel.
func (h *LogHook) Entry(maxlevel log.Level) *log.Entry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, e := range h.Entries {
		if e.Level <= maxlevel {
			return &e
		}
	}
	return nil
}

// Fire implements the logrus Hook interface
func (h *LogHook) Fire(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Entries = append(h.Entries, *e)
	return nil
}

// Levels implements the logrus Hook interface
func (h *LogHook) Levels() []log.Level {
	return log.AllLevels
}

// Reset clears the history of log entries in this LogHook instance.
func (h *LogHook) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Entries = make([]log.Entry, 0)
}
