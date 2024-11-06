package logger

import (
	"github.com/sethvargo/go-githubactions"
)

type actionLogger struct{}

var _ Logger = &actionLogger{}

// Debugf implements Logger.
func (a *actionLogger) Debugf(msg string, args ...any) {
	githubactions.Debugf(msg, args...)
}

// Errorf implements Logger.
func (a *actionLogger) Errorf(msg string, args ...any) {
	githubactions.Errorf(msg, args...)
}

// Fatalf implements Logger.
func (a *actionLogger) Fatalf(msg string, args ...any) {
	githubactions.Fatalf(msg, args...)
}

// Infof implements Logger.
func (a *actionLogger) Infof(msg string, args ...any) {
	githubactions.Infof(msg, args...)
}

// Noticef implements Logger.
func (a *actionLogger) Noticef(msg string, args ...any) {
	githubactions.Noticef(msg, args...)
}

// Warningf implements Logger.
func (a *actionLogger) Warningf(msg string, args ...any) {
	githubactions.Warningf(msg, args...)
}

func newActionLogger() Logger {
	return &actionLogger{}
}
