package logger

type noopLogger struct{}

var _ Logger = &noopLogger{}

// Debugf implements Logger.
func (*noopLogger) Debugf(_ string, _ ...any) {}

// Errorf implements Logger.
func (*noopLogger) Errorf(_ string, _ ...any) {}

// Fatalf implements Logger.
func (*noopLogger) Fatalf(_ string, _ ...any) {}

// Infof implements Logger.
func (*noopLogger) Infof(_ string, _ ...any) {}

// Noticef implements Logger.
func (*noopLogger) Noticef(_ string, _ ...any) {}

// Warningf implements Logger.
func (*noopLogger) Warningf(_ string, _ ...any) {}

func newNoopLogger() Logger {
	return &noopLogger{}
}
