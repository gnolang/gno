package log

import (
	"fmt"
	"io"

	"github.com/gnolang/gno/pkgs/colors"
)

const (
	logKeyLevel = ".level"
	logKeyMsg   = ".msg"
)

type tmLogger struct {
	level   LogLevel
	colorFn func(keyvals ...interface{}) colors.Color
	writer  io.Writer
}

var _ Logger = (*tmLogger)(nil)

func NewTMLogger(w io.Writer) Logger {
	// Color by level value
	colorFn := func(keyvals ...interface{}) colors.Color {
		if keyvals[0] != logKeyLevel {
			panic(fmt.Sprintf("expected level key to be first, got %v", keyvals[0]))
		}
		switch keyvals[1].(LogLevel) {
		case LevelDebug:
			return colors.Gray
		case LevelError:
			return colors.Red
		default:
			return colors.None
		}
	}
	return &tmLogger{
		level:   LevelDebug,
		colorFn: colorFn,
		writer:  w,
	}
}

// NewTMLoggerWithColorFn allows you to provide your own color function.
func NewTMLoggerWithColorFn(w io.Writer, colorFn func(keyvals ...interface{}) colors.Color) Logger {
	return &tmLogger{
		level:   LevelDebug,
		colorFn: colorFn,
		writer:  w,
	}
}

// Debug logs a message at level Debug.
func (l *tmLogger) Debug(msg string, keyvals ...interface{}) {
	if l.level <= LevelDebug {
		writeLog(l.writer, LevelDebug, l.colorFn, msg, keyvals...)
	}
}

// Info logs a message at level Info.
func (l *tmLogger) Info(msg string, keyvals ...interface{}) {
	if l.level <= LevelInfo {
		writeLog(l.writer, LevelInfo, l.colorFn, msg, keyvals...)
	}
}

// Error logs a message at level Error.
func (l *tmLogger) Error(msg string, keyvals ...interface{}) {
	if l.level <= LevelError {
		writeLog(l.writer, LevelError, l.colorFn, msg, keyvals...)
	}
}

// With returns a new contextual logger with keyvals prepended to those passed
// to calls to Info, Debug or Error.
func (l *tmLogger) With(keyvals ...interface{}) Logger {
	return newWithLogger(l, keyvals)
}

//----------------------------------------

type withLogger struct {
	base    Logger
	keyvals []interface{}
}

var _ Logger = (*withLogger)(nil)

func newWithLogger(base Logger, keyvals []interface{}) *withLogger {
	return &withLogger{
		base:    base,
		keyvals: keyvals,
	}
}

func (l *withLogger) Debug(msg string, keyvals ...interface{}) {
	keyvals = append(keyvals, l.keyvals)
	l.base.Debug(msg, keyvals)
}

func (l *withLogger) Info(msg string, keyvals ...interface{}) {
	keyvals = append(keyvals, l.keyvals)
	l.base.Info(msg, keyvals)
}

func (l *withLogger) Error(msg string, keyvals ...interface{}) {
	keyvals = append(keyvals, l.keyvals)
	l.base.Error(msg, keyvals)
}

func (l *withLogger) With(keyvals ...interface{}) Logger {
	return newWithLogger(l, keyvals)
}

//----------------------------------------

func writeLog(w io.Writer, level LogLevel, colorFn func(keyvals ...interface{}) colors.Color, msg string, keyvals ...interface{}) {
	keyvals = append([]interface{}{
		logKeyLevel, level,
		logKeyMsg, msg,
	}, keyvals...)
	color := colorFn(keyvals...)
	fmt.Fprintln(w, color(keyvals...))
}
