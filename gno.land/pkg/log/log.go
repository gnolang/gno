package log

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"

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

func NewLogger(w io.Writer, level zapcore.Level) (*slog.Logger, error) {
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

	defer zapLogger.Sync()

	handler := zapslog.NewHandler(
		zapLogger.Core(),
		&zapslog.HandlerOptions{
			LoggerName: "gno.land",
		},
	)

	return slog.New(handler), nil
}
