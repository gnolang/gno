package logger

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*Noop)(nil)

type Noop struct{}

func (Noop) Enabled(context.Context, slog.Level) bool { return false }

func (n Noop) Handle(context.Context, slog.Record) error { return nil }

func (n Noop) WithAttrs(attrs []slog.Attr) slog.Handler { return n }

func (n Noop) WithGroup(name string) slog.Handler { return n }
