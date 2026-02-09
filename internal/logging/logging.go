package logging

import (
	"context"
	"log/slog"
	"os"
)

const (
	KeyRequestID = "request_id"
	KeyTraceID   = "trace_id"
	KeyComponent = "component"
)

func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func WithRequest(ctx context.Context, logger *slog.Logger, requestID, traceID, component string) (*slog.Logger, context.Context) {
	logger = logger.With(
		slog.String(KeyRequestID, requestID),
		slog.String(KeyTraceID, traceID),
		slog.String(KeyComponent, component),
	)
	return logger, context.WithValue(ctx, loggerKey{}, logger)
}

type loggerKey struct{}

func FromContext(ctx context.Context, fallback *slog.Logger) *slog.Logger {
	if ctx == nil {
		return fallback
	}
	if logger, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return logger
	}
	return fallback
}
