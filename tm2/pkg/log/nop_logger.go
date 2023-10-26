package log

import (
	"context"
	"log/slog"
)

type NoopHandler struct{}

func NewNoopHandler() *NoopHandler {
	return &NoopHandler{}
}

func (n *NoopHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (n *NoopHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (n *NoopHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return n
}

func (n *NoopHandler) WithGroup(_ string) slog.Handler {
	return n
}
