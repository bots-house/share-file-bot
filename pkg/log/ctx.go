package log

import (
	"context"

	"github.com/go-kit/kit/log"
)

var noop = log.NewNopLogger()

type ctxKey int8

const (
	loggerCtxKey ctxKey = iota
)

// WithLogger set's context logger
func WithLogger(ctx context.Context, logger log.Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

// Logger returns context logger.
func Logger(ctx context.Context) log.Logger {
	v := ctx.Value(loggerCtxKey)
	if v != nil {
		return v.(log.Logger)
	}
	return noop
}
