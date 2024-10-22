package logger

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type FmtLoggerMessage interface {
	Fmt(message, name string) string
}

type defaultFmt struct {}

func (_ defaultFmt) Fmt(message, name string) string {
	return fmt.Sprintf("%s %s", strings.ToUpper(name), message)
}

type SlogAdapter struct {
	slogger    *slog.Logger
	fmtMessage FmtLoggerMessage
}

func (s *SlogAdapter) Enabled(level zapcore.Level) bool {
	return s.slogger.Enabled(context.Background(), zapToSlogLevel(level))
}

func (s *SlogAdapter) With(fields []zapcore.Field) zapcore.Core {
	return &SlogAdapter{slogger: s.slogger.With(convertZapFieldsToSloggerArgs(fields)...)}
}

func (s *SlogAdapter) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if s.Enabled(entry.Level) {
		return ce.AddCore(entry, s)
	}
	return ce
}

func (s *SlogAdapter) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	level := zapToSlogLevel(entry.Level)
	args := convertZapFieldsToSloggerArgs(fields)
	s.slogger.Log(context.Background(), level, s.fmtMessage.Fmt(entry.Message, entry.LoggerName), args...)
	// if entry.Level == zapcore.FatalLevel {
	//        os.Exit(1)
	// }
	return nil
}

// Sync flushes any buffered log entries from zap, not needed for slog.Logger
func (s *SlogAdapter) Sync() error {
	return nil
}

func convertZapFieldsToSloggerArgs(fields []zapcore.Field) []any {
	attrs := make([]any, len(fields)*2)
	i := 0

	for _, field := range fields {
		attrs[i] = field.Key
		i++
		attrs[i] = getFiledValue(field)
		i++
	}
	return attrs
}

func getFiledValue(f zapcore.Field) any {
	if f.Interface != nil {
		return f.Interface
	} else if f.String != "" {
		return f.String
	} else {
		return f.Integer
	}
}

func zapToSlogLevel(level zapcore.Level) slog.Level {
	switch level {
	case zapcore.DebugLevel:
		return slog.LevelDebug
	case zapcore.InfoLevel:
		return slog.LevelInfo
	case zapcore.WarnLevel:
		return slog.LevelWarn
	case zapcore.ErrorLevel:
		return slog.LevelError
	case zapcore.FatalLevel:
		return slog.LevelError // No Fatal level in slog
	default:
		return slog.LevelInfo
	}
}

// creates a zap.Logger that uses slog.Logger internally
func NewZapLoggerWithSlog(slogger *slog.Logger, fmt FmtLoggerMessage) *zap.Logger {
	core := &SlogAdapter{slogger: slogger, fmtMessage: defaultFmt{}}
	if fmt != nil {
		core.fmtMessage = fmt
	}

	return zap.New(core)
}

