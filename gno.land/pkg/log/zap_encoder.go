package log

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

func stableWidthCapitalLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%-5s", l.CapitalString()))
}

func stableWidthNameEncoder(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("%-18s", loggerName))
}

//nolint:varcheck,deadcode // we don't care if it's unused
const (
	black uint8 = iota + 30
	red
	green
	yellow
	blue
	magenta
	cyan
	white
)

func stableWidthCapitalColorLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", magenta, "DEBUG"))
	case zapcore.InfoLevel:
		enc.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", blue, "INFO "))
	case zapcore.WarnLevel:
		enc.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", yellow, "WARN "))
	case zapcore.ErrorLevel:
		enc.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", red, "ERROR"))
	case zapcore.DPanicLevel:
		enc.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", red, "DPANIC"))
	case zapcore.PanicLevel:
		enc.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", red, "PANIC"))
	case zapcore.FatalLevel:
		enc.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", red, "FATAL"))
	default:
		enc.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", red, l.CapitalString()))
	}
}
