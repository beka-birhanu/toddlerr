package logger

import "context"

// Logger is a structured, leveled logger.
type Logger interface {
	// With returns a child logger sharing the same underlying core, with
	// fields attached to every subsequent log call. The child does not own
	// any writers; closing it is a no-op, only the parent's Close matters.
	With(fields ...Field) Logger
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	// Fatal logs at fatal level, then terminates the process via os.Exit(1).
	Fatal(ctx context.Context, msg string, fields ...Field)
	// Close flushes and closes any writers registered via
	// WithCustomWriter. Safe to call on the root logger only; With()
	// children silently no-op.
	Close() error
}

// Field is a single structured log key/value pair.
type Field struct {
	Key string
	Val any
}
