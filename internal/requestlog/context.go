package requestlog

import (
	"context"

	"go.uber.org/zap"
)

const (
	RequestIDHeader = "X-Request-Id"
	RequestIDKey    = "request_id"
)

type contextKey string

const (
	loggerContextKey    contextKey = "logger"
	requestIDContextKey contextKey = "request_id"
)

func WithLogger(ctx context.Context, log *zap.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, loggerContextKey, log)
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func FromContext(ctx context.Context) *zap.Logger {
	if ctx != nil {
		if log, ok := ctx.Value(loggerContextKey).(*zap.Logger); ok && log != nil {
			return log
		}
	}
	return zap.NewNop()
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	requestID, _ := ctx.Value(requestIDContextKey).(string)
	return requestID
}
