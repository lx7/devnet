package turn

import (
	pionlog "github.com/pion/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LoggerFactory exists to provide compatibility with the pion LoggerFactory
// interface that is required for the turn module
type LoggerFactory struct {
}

func (lf LoggerFactory) NewLogger(scope string) pionlog.LeveledLogger {
	return LeveledLogger{
		logger: log.With().Str("scope", scope).Logger(),
	}
}

// LeveledLogger exists to provide compatibility with the pion LeveledLogger
// interface that is required for the turn module
type LeveledLogger struct {
	logger zerolog.Logger
}

func (l LeveledLogger) Trace(msg string) {
	l.logger.Trace().Msg(msg)
}

func (l LeveledLogger) Tracef(format string, args ...interface{}) {
	l.logger.Trace().Msgf(format, args...)
}

func (l LeveledLogger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

func (l LeveledLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l LeveledLogger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

func (l LeveledLogger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l LeveledLogger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

func (l LeveledLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l LeveledLogger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

func (l LeveledLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}
