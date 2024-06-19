package log

// Format is the log format
type Format string

const (
	JSONFormat    Format = "json"
	ConsoleFormat Format = "console"
	TestingFormat Format = "testing"
)

func (f Format) String() string {
	return string(f)
}
