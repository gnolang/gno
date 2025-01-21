package logger

import (
	"fmt"
	"log/slog"
	"os"
)

type termLogger struct{}

var _ Logger = &termLogger{}

// Debugf implements Logger.
func (s *termLogger) Debugf(msg string, args ...any) {
	msg = fmt.Sprintf("%s\n", msg)
	slog.Debug(fmt.Sprintf(msg, args...))
}

// Errorf implements Logger.
func (s *termLogger) Errorf(msg string, args ...any) {
	msg = fmt.Sprintf("%s\n", msg)
	slog.Error(fmt.Sprintf(msg, args...))
}

// Fatalf implements Logger.
func (s *termLogger) Fatalf(msg string, args ...any) {
	s.Errorf(msg, args...)
	os.Exit(1)
}

// Infof implements Logger.
func (s *termLogger) Infof(msg string, args ...any) {
	msg = fmt.Sprintf("%s\n", msg)
	slog.Info(fmt.Sprintf(msg, args...))
}

// Noticef implements Logger.
func (s *termLogger) Noticef(msg string, args ...any) {
	// Alias to info on terminal since notice level only exists on GitHub Actions.
	s.Infof(msg, args...)
}

// Warningf implements Logger.
func (s *termLogger) Warningf(msg string, args ...any) {
	msg = fmt.Sprintf("%s\n", msg)
	slog.Warn(fmt.Sprintf(msg, args...))
}

func newTermLogger(verbose bool) Logger {
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	return &termLogger{}
}
