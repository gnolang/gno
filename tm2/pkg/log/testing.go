package log

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// NewTestingLogger returns a new testing logger
func NewTestingLogger(t *testing.T) *slog.Logger {
	t.Helper()

	// Parse the environment vars
	envLevel := os.Getenv("LOG_LEVEL")
	envPath := os.Getenv("LOG_PATH_DIR")

	if !testing.Verbose() && envLevel == "" && envPath == "" {
		return NewNoopLogger()
	}

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
		// Create the top-level log directory
		if err := os.Mkdir(envPath, 0o755); err != nil && !os.IsExist(err) {
			t.Fatalf("Failed to create log directory: %v", err)
		}

		logName := fmt.Sprintf(
			"%s-%d.log",
			strings.ReplaceAll(t.Name(), "/", "_"), // unique test name
			time.Now().Unix(),                      // unique test timestamp
		)
		logPath := filepath.Join(envPath, logName)

		logFile, err := os.Create(logPath)
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
