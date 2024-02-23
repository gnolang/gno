package log

import (
	"context"
	"log/slog"
)

// NewNoopLogger returns a new no-op logger
func NewNoopLogger() *slog.Logger {
	return slog.New(newNoopHandler())
}

type noopHandler struct{}

func newNoopHandler() *noopHandler {
	return &noopHandler{}
}

func (n *noopHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (n *noopHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (n *noopHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return n
}

func (n *noopHandler) WithGroup(_ string) slog.Handler {
	return n
}
