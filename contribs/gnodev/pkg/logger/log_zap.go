package logger

import (
	"io"
	"log/slog"

	"github.com/gnolang/gno/gno.land/pkg/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(w io.Writer, slevel slog.Level) *zap.Logger {
	// Build encoder config
	consoleConfig := zap.NewDevelopmentEncoderConfig()
	consoleConfig.EncodeCaller = zapcore.FullCallerEncoder
	consoleConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	consoleConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleConfig.EncodeName = zapcore.FullNameEncoder

	// Build encoder
	enc := zapcore.NewConsoleEncoder(consoleConfig)

	var level zapcore.Level
	switch slevel {
	case slog.LevelDebug:
		level = zapcore.DebugLevel
	case slog.LevelError:
		level = zapcore.ErrorLevel
	case slog.LevelInfo:
		level = zapcore.InfoLevel
	case slog.LevelWarn:
		level = zapcore.WarnLevel
	default:
		panic("invalid slog level")
	}

	return log.NewZapLogger(enc, w, level)
}
