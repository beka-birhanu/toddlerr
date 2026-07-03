package logger

import (
	"io"
	"os"

	"go.uber.org/zap/zapcore"
)

// Option configures a Logger built by NewLogger.
type Option func(*logger)

// OptNoop discards all output. Overrides every writer/level option.
func OptNoop() Option {
	return func(logger *logger) {
		logger.noopLogger = true
	}
}

// MaskEnabled turns on struct field masking (fields tagged `mask:""`).
// Masking is off by default.
func MaskEnabled() Option {
	return func(logger *logger) {
		logger.maskEnabled = true
	}
}

// WithStdout adds os.Stdout as a log destination. Applied automatically if
// no writer option is given at all.
func WithStdout() Option {
	return func(logger *logger) {
		logger.writers = append(logger.writers, os.Stdout)
	}
}

// WithCustomWriter adds writer as a log destination; it is closed on
// Logger.Close. A nil writer is a no-op.
func WithCustomWriter(writer io.WriteCloser) Option {
	if writer == nil {
		return func(logger *logger) {}
	}
	return func(logger *logger) {
		logger.writers = append(logger.writers, writer)
		logger.closer = append(logger.closer, writer)
	}
}

// WithLevel sets the minimum level logged (default: zapcore.InfoLevel).
func WithLevel(level zapcore.Level) Option {
	return func(logger *logger) {
		logger.level = level
	}
}
