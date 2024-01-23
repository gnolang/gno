package log

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jaekwon/testify/require"
	"golang.org/x/exp/slog"
)

// NewTestingLogger returns a new testing logger
func NewTestingLogger(t *testing.T) *slog.Logger {
	t.Helper()

	if !testing.Verbose() {
		return NewNoopLogger()
	}

	// Parse the environment vars
	envLevel := os.Getenv("LOG_LEVEL")
	envPath := os.Getenv("LOG_PATH")

	// Default logger config
	logLevel := slog.LevelError
	logOutput := os.Stdout

	// Set the logger level, if any
	switch strings.ToLower(envLevel) {
	case "info":
		logLevel = slog.LevelInfo
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	}

	// Check if the log output needs to be a file
	if envPath != "" {
		logFile, err := os.Create(
			fmt.Sprintf(
				"%s_%s",
				envPath,
				t.Name(),
			),
		)
		require.NoError(t, err)

		t.Cleanup(func() {
			_ = logFile.Close()
		})

		logOutput = logFile
	}

	// Create the log handler
	logHandler := slog.NewTextHandler(
		logOutput,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     logLevel,
		},
	)

	return slog.New(logHandler)
}
