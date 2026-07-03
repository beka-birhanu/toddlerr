package logger

import (
	"context"
	"errors"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logger struct {
	writers     []io.Writer
	maskEnabled bool
	noopLogger  bool
	closer      []io.Closer

	zapLogger *zap.Logger
	level     zapcore.Level
}

var _ Logger = (*logger)(nil)

// NewLogger builds a Logger. With no options it writes info-and-above JSON
// logs to stdout, unmasked. Store the result in a package-level var and reuse
// it — see Logger.With for per-request children.
func NewLogger(opts ...Option) Logger {
	l := &logger{
		writers:     make([]io.Writer, 0),
		maskEnabled: false,
		noopLogger:  false,
		closer:      make([]io.Closer, 0),
		level:       zapcore.InfoLevel,
	}

	for _, o := range opts {
		o(l)
	}

	l.zapLogger = newZapLogger(l.level, l.writers...)

	if len(l.writers) <= 0 {
		l.zapLogger = newZapLogger(l.level, zapcore.AddSync(os.Stdout))
	}

	if l.noopLogger {
		l.zapLogger = zap.NewNop()
	}

	return l
}

func (d *logger) With(fields ...Field) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		zapFields = append(zapFields, formatLog(f.Key, f.Val, d.maskEnabled))
	}
	return &logger{
		zapLogger:   d.zapLogger.With(zapFields...),
		maskEnabled: d.maskEnabled,
		level:       d.level,
	}
}

func (d *logger) Close() error {
	if d.closer == nil {
		return nil
	}

	var errs []error
	for _, closer := range d.closer {
		if closer == nil {
			continue
		}
		if e := closer.Close(); e != nil {
			errs = append(errs, e)
		}
	}

	return errors.Join(errs...)
}

func (d *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	zapLogs := []zap.Field{zap.String("logType", LogTypeSYS), zap.String("level", "debug")}
	zapLogs = append(zapLogs, formatLogs(msg, d.maskEnabled, fields...)...)
	d.zapLogger.Debug(separator, zapLogs...)
}

func (d *logger) Info(ctx context.Context, msg string, fields ...Field) {
	zapLogs := []zap.Field{zap.String("logType", LogTypeSYS), zap.String("level", "info")}
	zapLogs = append(zapLogs, formatLogs(msg, d.maskEnabled, fields...)...)
	d.zapLogger.Info(separator, zapLogs...)
}

func (d *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	zapLogs := []zap.Field{zap.String("logType", LogTypeSYS), zap.String("level", "warn")}
	zapLogs = append(zapLogs, formatLogs(msg, d.maskEnabled, fields...)...)
	d.zapLogger.Warn(separator, zapLogs...)
}

func (d *logger) Error(ctx context.Context, msg string, fields ...Field) {
	zapLogs := []zap.Field{zap.String("logType", LogTypeSYS), zap.String("level", "error")}
	zapLogs = append(zapLogs, formatLogs(msg, d.maskEnabled, fields...)...)
	d.zapLogger.Error(separator, zapLogs...)
}

func (d *logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	zapLogs := []zap.Field{zap.String("logType", LogTypeSYS), zap.String("level", "fatal")}
	zapLogs = append(zapLogs, formatLogs(msg, d.maskEnabled, fields...)...)
	d.zapLogger.Fatal(separator, zapLogs...)
}
