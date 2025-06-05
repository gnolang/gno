package noop

// Logger is the nil (noop) logger
type Logger struct{}

// New creates a nil logger
func New() *Logger {
	return &Logger{}
}

func (l Logger) Info(_ string, _ ...interface{}) {}

func (l Logger) Debug(_ string, _ ...interface{}) {}

func (l Logger) Error(_ string, _ ...interface{}) {}
