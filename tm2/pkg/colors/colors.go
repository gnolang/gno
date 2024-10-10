package colors

import (
	"fmt"
	"strings"
)

type Color func(args ...interface{}) string

const (
	ANSIReset      = "\x1b[0m"
	ANSIBright     = "\x1b[1m"
	ANSIDim        = "\x1b[2m"
	ANSIUnderscore = "\x1b[4m"
	ANSIBlink      = "\x1b[5m"
	ANSIReverse    = "\x1b[7m"
	ANSIHidden     = "\x1b[8m"

	ANSIFgBlack   = "\x1b[30m"
	ANSIFgRed     = "\x1b[31m"
	ANSIFgGreen   = "\x1b[32m"
	ANSIFgYellow  = "\x1b[33m"
	ANSIFgBlue    = "\x1b[34m"
	ANSIFgMagenta = "\x1b[35m"
	ANSIFgCyan    = "\x1b[36m"
	ANSIFgWhite   = "\x1b[37m"

	ANSIFgGray = "\x1b[90m" // bright black

	ANSIBgBlack   = "\x1b[40m"
	ANSIBgRed     = "\x1b[41m"
	ANSIBgGreen   = "\x1b[42m"
	ANSIBgYellow  = "\x1b[43m"
	ANSIBgBlue    = "\x1b[44m"
	ANSIBgMagenta = "\x1b[45m"
	ANSIBgCyan    = "\x1b[46m"
	ANSIBgWhite   = "\x1b[47m"
)

// color the string s with color 'color'
// unless s is already colored
func treat(s string, color string) string {
	if len(s) > 2 && s[:2] == "\x1b[" {
		return s
	}
	return color + s + ANSIReset
}

func treatAll(color string, args ...interface{}) string {
	parts := make([]string, len(args))

	for i, arg := range args {
		parts[i] = treat(fmt.Sprintf("%v", arg), color)
	}

	return strings.Join(parts, "")
}

func None(args ...interface{}) string {
	return treatAll(ANSIReset, args...)
}

func Black(args ...interface{}) string {
	return treatAll(ANSIFgBlack, args...)
}

func Red(args ...interface{}) string {
	return treatAll(ANSIFgRed, args...)
}

func Green(args ...interface{}) string {
	return treatAll(ANSIFgGreen, args...)
}

func Yellow(args ...interface{}) string {
	return treatAll(ANSIFgYellow, args...)
}

func Blue(args ...interface{}) string {
	return treatAll(ANSIFgBlue, args...)
}

func Magenta(args ...interface{}) string {
	return treatAll(ANSIFgMagenta, args...)
}

func Cyan(args ...interface{}) string {
	return treatAll(ANSIFgCyan, args...)
}

func White(args ...interface{}) string {
	return treatAll(ANSIFgWhite, args...)
}

func Gray(args ...interface{}) string {
	return treatAll(ANSIFgGray, args...)
}

// result may be 4 ASNSII chars longer than they should be to denote the
// elipses (...), and one for a trailing hex nibble in case the last byte is
// non-ascii.
// NOTE: it is annoying to try make this perfect and always fit within n, so we
// don't do this yet, but left as an exercise. :)
func ColoredBytesN(data []byte, n int, textColor, bytesColor func(...interface{}) string) string {
	_n := 0
	s := ""
	buf := ""         // buffer
	bufIsText := true // is buf text or hex
	for i, b := range data {
	RESTART:
		if 0x21 <= b && b < 0x7F {
			if !bufIsText {
				s += bytesColor(buf)
				buf = ""
				bufIsText = true
				goto RESTART
			}
			buf += string(b)
			_n += 1
			if n != 0 && _n >= n {
				if i == len(data)-1 {
					// done
					s += textColor(buf)
					buf = ""
				} else {
					s += textColor(buf) + "..."
					buf = ""
				}
				break
			}
		} else {
			if bufIsText {
				s += textColor(buf)
				buf = ""
				bufIsText = false
				goto RESTART
			}
			buf += fmt.Sprintf("%02X", b)
			_n += 2
			if n != 0 && _n >= n {
				if i == len(data)-1 {
					// done
					s += bytesColor(buf)
					buf = ""
				} else {
					s += bytesColor(buf) + "..."
					buf = ""
				}
				break
			}
		}
	}
	if buf != "" {
		if bufIsText {
			s += textColor(buf)
			buf = ""
		} else {
			s += bytesColor(buf)
			buf = ""
		}
	}
	return s
}

func DefaultColoredBytesN(data []byte, n int) string {
	return ColoredBytesN(data, n, Blue, Green)
}

func ColoredBytes(data []byte, textColor, bytesColor func(...interface{}) string) string {
	return ColoredBytesN(data, 0, textColor, bytesColor)
}

func DefaultColoredBytes(data []byte) string {
	return ColoredBytes(data, Blue, Green)
}

func ColoredBytesOnlyAscii(data []byte, textColor func(...interface{}) string) string {
	s := ""
	for _, b := range data {
		if 0x21 <= b && b < 0x7F {
			s += textColor(string(b))
		} else {
			s += string(b)
		}
	}
	return s
}
