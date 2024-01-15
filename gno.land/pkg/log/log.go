package log

import (
	"fmt"
	"io"
	"net/url"

	"golang.org/x/exp/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
	"moul.io/zapconfig"
)

const customWriterKey = "gno"

type customWriter struct {
	io.Writer
}

func (cw customWriter) Close() error {
	return nil
}

func (cw customWriter) Sync() error {
	return nil
}

// NewZapLogger creates a new zap logger instance, for the given level and writer
func NewZapLogger(w io.Writer, level zapcore.Level) (*zap.Logger, error) {
	if err := zap.RegisterSink(
		customWriterKey,
		func(u *url.URL) (zap.Sink, error) {
			return customWriter{w}, nil
		}); err != nil {
		return nil, fmt.Errorf("unable to register sink, %w", err)
	}

	zapLogger, err := zapconfig.New().
		SetOutputPath(fmt.Sprintf("%s:", customWriterKey)).
		EnableStacktrace().
		SetLevel(level).
		SetPreset("console").
		Build()
	if err != nil {
		return nil, fmt.Errorf("unable to build logger, %w", err)
	}

	return zapLogger, nil
}

// ZapLoggerToSlog wraps the given zap logger to an log/slog Logger
func ZapLoggerToSlog(logger *zap.Logger) *slog.Logger {
	return slog.New(zapslog.NewHandler(logger.Core()))
}
