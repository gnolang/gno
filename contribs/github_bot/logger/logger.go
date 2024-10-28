package logger

import (
	"os"
)

// All Logger methods follow the standard fmt.Printf convention
type Logger interface {
	// Debugf prints a debug-level message
	Debugf(msg string, args ...any)

	// Noticef prints a notice-level message
	Noticef(msg string, args ...any)

	// Warningf prints a warning-level message
	Warningf(msg string, args ...any)

	// Errorf prints a error-level message
	Errorf(msg string, args ...any)

	// Fatalf prints a error-level message and exits
	Fatalf(msg string, args ...any)

	// Infof prints message to stdout without any level annotations
	Infof(msg string, args ...any)
}

func NewLogger(verbose bool) Logger {
	if _, isAction := os.LookupEnv("GITHUB_ACTION"); isAction {
		return newActionLogger()
	}

	return newTermLogger(verbose)
}
