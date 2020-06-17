package log

import (
	"context"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/log/term"
)

func NewLogger(debug bool, colors bool) log.Logger {
	var logger log.Logger

	writer := log.NewSyncWriter(os.Stdout)
	if colors {
		logger = term.NewLogger(writer, log.NewLogfmtLogger, loggerColorFn)
	} else {
		logger = log.NewLogfmtLogger(writer)
	}

	logger = log.With(logger, "ts", log.TimestampFormat(time.Now, "15:04:05"))

	if !debug {
		logger = level.NewFilter(logger, level.AllowInfo())
	} else {
		logger = level.NewFilter(logger, level.AllowDebug())
	}

	return logger
}

// With creates a context with new key values.
func With(ctx context.Context, kvs ...interface{}) context.Context {
	logger := Logger(ctx)
	logger = log.With(logger, kvs...)

	return WithLogger(ctx, logger)
}

// WithPrefix create a context with prefix.
func WithPrefix(ctx context.Context, kvs ...interface{}) context.Context {
	logger := Logger(ctx)
	logger = log.WithPrefix(logger, kvs...)

	return WithLogger(ctx, logger)
}

// Log message.
func Log(ctx context.Context, msg string, kvs ...interface{}) {
	kvs = append([]interface{}{
		"msg",
		msg,
	}, kvs...)

	_ = Logger(ctx).Log(kvs...)
}

// Debug message
func Debug(ctx context.Context, msg string, kvs ...interface{}) {
	kvs = append([]interface{}{
		"msg",
		msg,
	}, kvs...)

	_ = level.Debug(Logger(ctx)).Log(kvs...)
}

// Info message
func Info(ctx context.Context, msg string, kvs ...interface{}) {
	kvs = append([]interface{}{
		"msg",
		msg,
	}, kvs...)

	_ = level.Info(Logger(ctx)).Log(kvs...)
}

// Error message
func Error(ctx context.Context, msg string, kvs ...interface{}) {
	kvs = append([]interface{}{
		"msg",
		msg,
	}, kvs...)

	_ = level.Error(Logger(ctx)).Log(kvs...)
}

// Warn message
func Warn(ctx context.Context, msg string, kvs ...interface{}) {
	kvs = append([]interface{}{
		"msg",
		msg,
	}, kvs...)

	_ = level.Warn(Logger(ctx)).Log(kvs...)
}
