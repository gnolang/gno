package log

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

// NewZapLoggerFn is the zap logger init declaration
type NewZapLoggerFn func(w io.Writer, level zapcore.Level, opts ...zap.Option) *zap.Logger

// GetZapLoggerFn returns the appropriate init callback
// for the zap logger, given the requested format
func GetZapLoggerFn(format Format) NewZapLoggerFn {
	switch format {
	case JSONFormat:
		return NewZapJSONLogger
	case TestingFormat:
		return NewZapTestingLogger
	default:
		return NewZapConsoleLogger
	}
}

// InitializeZapLogger initializes the zap logger using the given format and log level,
// outputting to the given IO
func InitializeZapLogger(io io.WriteCloser, logLevel, logFormat string) (*zap.Logger, error) {
	// Initialize the log level
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("unable to parse log level, %w", err)
	}

	// Initialize the log format
	format := Format(strings.ToLower(logFormat))

	// Initialize the zap logger
	return GetZapLoggerFn(format)(io, level), nil
}

// NewZapJSONLogger creates a zap logger with a JSON encoder for production use.
func NewZapJSONLogger(w io.Writer, level zapcore.Level, opts ...zap.Option) *zap.Logger {
	// Build encoder config
	jsonConfig := zap.NewProductionEncoderConfig()

	// Build encoder
	enc := zapcore.NewJSONEncoder(jsonConfig)
	return NewZapLogger(enc, w, level, opts...)
}

// NewZapConsoleLogger creates a zap logger with a console encoder for development use.
func NewZapConsoleLogger(w io.Writer, level zapcore.Level, opts ...zap.Option) *zap.Logger {
	// Build encoder config
	consoleConfig := zap.NewDevelopmentEncoderConfig()
	consoleConfig.EncodeLevel = stableWidthCapitalColorLevelEncoder
	consoleConfig.EncodeName = stableWidthNameEncoder

	// Build encoder
	enc := zapcore.NewConsoleEncoder(consoleConfig)
	return NewZapLogger(enc, w, level, opts...)
}

// NewZapTestingLogger creates a zap logger with a console encoder optimized for testing.
func NewZapTestingLogger(w io.Writer, level zapcore.Level, opts ...zap.Option) *zap.Logger {
	// Build encoder config
	consoleConfig := zap.NewDevelopmentEncoderConfig()
	consoleConfig.TimeKey = ""
	consoleConfig.EncodeLevel = stableWidthCapitalLevelEncoder
	consoleConfig.EncodeName = stableWidthNameEncoder

	// Build encoder
	enc := zapcore.NewConsoleEncoder(consoleConfig)
	return NewZapLogger(enc, w, level, opts...)
}

// NewZapLogger creates a new zap logger instance, for the given level, writer and zap encoder.
func NewZapLogger(enc zapcore.Encoder, w io.Writer, level zapcore.Level, opts ...zap.Option) *zap.Logger {
	ws := zapcore.AddSync(w)

	// Create zap core
	core := zapcore.NewCore(enc, ws, zap.NewAtomicLevelAt(level))
	return zap.New(core, opts...)
}

// ZapLoggerToSlog wraps the given zap logger to an log/slog Logger
func ZapLoggerToSlog(logger *zap.Logger) *slog.Logger {
	return slog.New(zapslog.NewHandler(logger.Core()))
}
