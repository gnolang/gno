package log

import (
	"fmt"
	"io"
	"log/slog"
	"net/url"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

const customWriterKey = "tm2"

type customWriter struct {
	io.Writer
}

func (cw customWriter) Close() error {
	return nil
}

func (cw customWriter) Sync() error {
	return nil
}

func NewLogger(w io.Writer, level slog.Level) (*slog.Logger, error) {
	config := zap.NewDevelopmentConfig()

	switch level {
	case slog.LevelInfo:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case slog.LevelDebug:
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case slog.LevelError:
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case slog.LevelWarn:
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	}

	err := zap.RegisterSink(customWriterKey, func(u *url.URL) (zap.Sink, error) {
		return customWriter{w}, nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to register sink, %w", err)
	}

	config.OutputPaths = []string{
		fmt.Sprintf("%s:", customWriterKey),
	}

	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to build logger config, %w", err)
	}

	defer zapLogger.Sync()

	handler := zapslog.NewHandler(zapLogger.Core(), &zapslog.HandlerOptions{LoggerName: "tm2"})

	return slog.New(handler), nil
}
