package turn

import (
	"testing"

	"github.com/lx7/devnet/internal/testutil"
	pionlog "github.com/pion/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	// initialize log hook
	hook := &testutil.LogHook{}
	log.Logger = log.Hook(hook)

	lf := LoggerFactory{}

	// define cases
	tests := []struct {
		desc  string
		level zerolog.Level
		run   func(*testing.T, pionlog.LeveledLogger)
	}{
		{
			level: zerolog.TraceLevel,
			run: func(t *testing.T, l pionlog.LeveledLogger) {
				l.Trace("test")
			},
		},
		{
			level: zerolog.DebugLevel,
			run: func(t *testing.T, l pionlog.LeveledLogger) {
				l.Debug("test")
			},
		},
		{
			level: zerolog.InfoLevel,
			run: func(t *testing.T, l pionlog.LeveledLogger) {
				l.Info("test")
			},
		},
		{
			level: zerolog.WarnLevel,
			run: func(t *testing.T, l pionlog.LeveledLogger) {
				l.Warn("test")
			},
		},
		{
			level: zerolog.ErrorLevel,
			run: func(t *testing.T, l pionlog.LeveledLogger) {
				l.Error("test")
			},
		},
		{
			level: zerolog.TraceLevel,
			run: func(t *testing.T, l pionlog.LeveledLogger) {
				l.Tracef("t%s%s%s", "e", "s", "t")
			},
		},
		{
			level: zerolog.DebugLevel,
			run: func(t *testing.T, l pionlog.LeveledLogger) {
				l.Debugf("t%s%s%s", "e", "s", "t")
			},
		},
		{
			level: zerolog.InfoLevel,
			run: func(t *testing.T, l pionlog.LeveledLogger) {
				l.Infof("t%s%s%s", "e", "s", "t")
			},
		},
		{
			level: zerolog.WarnLevel,
			run: func(t *testing.T, l pionlog.LeveledLogger) {
				l.Warnf("t%s%s%s", "e", "s", "t")
			},
		},
		{
			level: zerolog.ErrorLevel,
			run: func(t *testing.T, l pionlog.LeveledLogger) {
				l.Errorf("t%s%s%s", "e", "s", "t")
			},
		},
	}

	// run tests
	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			l := lf.NewLogger(tt.level.String())
			tt.run(t, l)

			entry := hook.Entry(tt.level)
			assert.NotNil(t, entry, "log entry expexted")
			assert.Equal(t, entry.Msg, "test", "log msg should match")
			hook.Reset()
		})
	}
}
