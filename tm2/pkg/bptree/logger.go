package bptree

// Logger is the interface for tree logging.
type Logger interface {
	Info(msg string, keyVals ...any)
	Warn(msg string, keyVals ...any)
	Error(msg string, keyVals ...any)
	Debug(msg string, keyVals ...any)
}

type nopLogger struct{}

func (nopLogger) Info(string, ...any)  {}
func (nopLogger) Warn(string, ...any)  {}
func (nopLogger) Error(string, ...any) {}
func (nopLogger) Debug(string, ...any) {}

// NewNopLogger returns a Logger that discards all output.
func NewNopLogger() Logger { return nopLogger{} }
