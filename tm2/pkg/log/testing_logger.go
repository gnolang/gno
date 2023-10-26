package log

import (
	"io"
	"log/slog"
	"os"
	"testing"
)

// TestingLogger returns a TMLogger which writes to STDOUT if testing being run
// with the verbose (-v) flag, NopLogger otherwise.
//
// Note that the call to TestingLogger() must be made
// inside a test (not in the init func) because
// verbose flag only set at the time of testing.
func TestingLogger() *slog.Logger {
	return TestingLoggerWithOutput(os.Stdout)
}

// TestingLoggerWithOutput returns a TMLogger which writes to (w io.Writer) if testing being run
// with the verbose (-v) flag, NopLogger otherwise.
//
// Note that the call to TestingLoggerWithOutput(w io.Writer) must be made
// inside a test (not in the init func) because
// verbose flag only set at the time of testing.
func TestingLoggerWithOutput(w io.Writer) *slog.Logger {
	if testing.Verbose() {
		logger, _ := NewTMLogger(w, slog.LevelDebug)

		return logger
	}

	return slog.New(NewNoopHandler())
}
