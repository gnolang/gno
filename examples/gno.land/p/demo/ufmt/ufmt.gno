// Package ufmt provides utility functions for formatting strings, similarly to
// the Go package "fmt", of which only a subset is currently supported (hence
// the name µfmt - micro fmt). It includes functions like Printf, Sprintf,
// Fprintf, and Errorf.
// Supported formatting verbs are documented in the Sprintf function.
package ufmt

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

// buffer accumulates formatted output as a byte slice.
type buffer []byte

func (b *buffer) write(p []byte) {
	*b = append(*b, p...)
}

func (b *buffer) writeString(s string) {
	*b = append(*b, s...)
}

func (b *buffer) writeByte(c byte) {
	*b = append(*b, c)
}

func (b *buffer) writeRune(r rune) {
	*b = utf8.AppendRune(*b, r)
}

// printer holds state for formatting operations.
type printer struct {
	buf buffer
}

func newPrinter() *printer {
	return &printer{}
}

// Sprint formats using the default formats for its operands and returns the resulting string.
// Sprint writes the given arguments with spaces between arguments.
func Sprint(a ...any) string {
	p := newPrinter()
	p.doPrint(a)
	return string(p.buf)
}

// doPrint formats arguments using default formats and writes to printer's buffer.
// Spaces are added between arguments.
func (p *printer) doPrint(args []any) {
	for argNum, arg := range args {
		if argNum > 0 {
			p.buf.writeRune(' ')
		}

		switch v := arg.(type) {
		case string:
			p.buf.writeString(v)
		case (interface{ String() string }):
			p.buf.writeString(v.String())
		case error:
			p.buf.writeString(v.Error())
		case float64:
			p.buf.writeString(Sprintf("%f", v))
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			p.buf.writeString(Sprintf("%d", v))
		case bool:
			if v {
				p.buf.writeString("true")
			} else {
				p.buf.writeString("false")
			}
		case nil:
			p.buf.writeString("<nil>")
		default:
			p.buf.writeString("(unhandled)")
		}
	}
}

// doPrintln appends a newline after formatting arguments with doPrint.
func (p *printer) doPrintln(a []any) {
	p.doPrint(a)
	p.buf.writeByte('\n')
}

// Sprintf offers similar functionality to Go's fmt.Sprintf, or the sprintf
// equivalent available in many languages, including C/C++.
// The number of args passed must exactly match the arguments consumed by the format.
// A limited number of formatting verbs and features are currently supported.
//
// Supported verbs:
//
//	%s: Places a string value directly.
//	    If the value implements the interface interface{ String() string },
//	    the String() method is called to retrieve the value. Same about Error()
//	    string.
//	%c: Formats the character represented by Unicode code point
//	%d: Formats an integer value using package "strconv".
//	    Currently supports only uint, uint64, int, int64.
//	%f: Formats a float value, with a default precision of 6.
//	%e: Formats a float with scientific notation; 1.23456e+78
//	%E: Formats a float with scientific notation; 1.23456E+78
//	%F: The same as %f
//	%g: Formats a float value with %e for large exponents, and %f with full precision for smaller numbers
//	%G: Formats a float value with %G for large exponents, and %F with full precision for smaller numbers
//	%t: Formats a boolean value to "true" or "false".
//	%x: Formats an integer value as a hexadecimal string.
//	    Currently supports only uint8, []uint8, [32]uint8.
//	%c: Formats a rune value as a string.
//	    Currently supports only rune, int.
//	%q: Formats a string value as a quoted string.
//	%T: Formats the type of the value.
//	%v: Formats the value with a default representation appropriate for the value's type
//	    - nil: <nil>
//	    - bool: true/false
//	    - integers: base 10
//	    - float64: %g format
//	    - string: verbatim
//	    - types with String()/Error(): method result
//	    - others: (unhandled)
//	%%: Outputs a literal %. Does not consume an argument.
//
// Unsupported verbs or type mismatches produce error strings like "%!d(string=foo)".
func Sprintf(format string, a ...any) string {
	p := newPrinter()
	p.doPrintf(format, a)
	return string(p.buf)
}

// doPrintf parses the format string and writes formatted arguments to the buffer.
func (p *printer) doPrintf(format string, args []any) {
	sTor := []rune(format)
	end := len(sTor)
	argNum := 0
	argLen := len(args)

	for i := 0; i < end; {
		isLast := i == end-1
		c := sTor[i]

		if isLast || c != '%' {
			// we don't check for invalid format like a one ending with "%"
			p.buf.writeRune(c)
			i++
			continue
		}

		length := -1
		precision := -1
		i++ // skip '%'

		digits := func() string {
			start := i
			for i < end && sTor[i] >= '0' && sTor[i] <= '9' {
				i++
			}
			if i > start {
				return string(sTor[start:i])
			}
			return ""
		}

		if l := digits(); l != "" {
			var err error
			length, err = strconv.Atoi(l)
			if err != nil {
				panic("ufmt: invalid length specification")
			}
		}

		if i < end && sTor[i] == '.' {
			i++ // skip '.'
			if l := digits(); l != "" {
				var err error
				precision, err = strconv.Atoi(l)
				if err != nil {
					panic("ufmt: invalid precision specification")
				}
			}
		}

		if i >= end {
			panic("ufmt: invalid format string")
		}

		verb := sTor[i]
		if verb == '%' {
			p.buf.writeRune('%')
			i++
			continue
		}

		if argNum >= argLen {
			panic("ufmt: not enough arguments")
		}
		arg := args[argNum]
		argNum++

		switch verb {
		case 'v':
			writeValue(p, verb, arg)
		case 's':
			writeStringWithLength(p, verb, arg, length)
		case 'c':
			writeChar(p, verb, arg)
		case 'd':
			writeInt(p, verb, arg)
		case 'e', 'E', 'f', 'F', 'g', 'G':
			writeFloatWithPrecision(p, verb, arg, precision)
		case 't':
			writeBool(p, verb, arg)
		case 'x':
			writeHex(p, verb, arg)
		case 'q':
			writeQuotedString(p, verb, arg)
		case 'T':
			writeType(p, arg)
		// % handled before, as it does not consume an argument
		default:
			p.buf.writeString("(unhandled verb: %" + string(verb) + ")")
		}

		i++
	}

	if argNum < argLen {
		panic("ufmt: too many arguments")
	}
}

// writeValue handles %v formatting
func writeValue(p *printer, verb rune, arg any) {
	switch v := arg.(type) {
	case nil:
		p.buf.writeString("<nil>")
	case bool:
		writeBool(p, verb, v)
	case int:
		p.buf.writeString(strconv.Itoa(v))
	case int8:
		p.buf.writeString(strconv.Itoa(int(v)))
	case int16:
		p.buf.writeString(strconv.Itoa(int(v)))
	case int32:
		p.buf.writeString(strconv.Itoa(int(v)))
	case int64:
		p.buf.writeString(strconv.Itoa(int(v)))
	case uint:
		p.buf.writeString(strconv.FormatUint(uint64(v), 10))
	case uint8:
		p.buf.writeString(strconv.FormatUint(uint64(v), 10))
	case uint16:
		p.buf.writeString(strconv.FormatUint(uint64(v), 10))
	case uint32:
		p.buf.writeString(strconv.FormatUint(uint64(v), 10))
	case uint64:
		p.buf.writeString(strconv.FormatUint(v, 10))
	case float64:
		p.buf.writeString(strconv.FormatFloat(v, 'g', -1, 64))
	case string:
		p.buf.writeString(v)
	case []byte:
		p.buf.write(v)
	case []rune:
		p.buf.writeString(string(v))
	case (interface{ String() string }):
		p.buf.writeString(v.String())
	case error:
		p.buf.writeString(v.Error())
	default:
		p.buf.writeString(fallback(verb, v))
	}
}

// writeStringWithLength handles %s formatting with length specification
func writeStringWithLength(p *printer, verb rune, arg any, length int) {
	var s string
	switch v := arg.(type) {
	case (interface{ String() string }):
		s = v.String()
	case error:
		s = v.Error()
	case string:
		s = v
	default:
		s = fallback(verb, v)
	}

	if length > 0 && len(s) < length {
		s = strings.Repeat(" ", length-len(s)) + s
	}
	p.buf.writeString(s)
}

// writeChar handles %c formatting
func writeChar(p *printer, verb rune, arg any) {
	switch v := arg.(type) {
	// rune is int32. Exclude overflowing numeric types and dups (byte, int32):
	case rune:
		p.buf.writeString(string(v))
	case int:
		p.buf.writeRune(rune(v))
	case int8:
		p.buf.writeRune(rune(v))
	case int16:
		p.buf.writeRune(rune(v))
	case uint:
		p.buf.writeRune(rune(v))
	case uint8:
		p.buf.writeRune(rune(v))
	case uint16:
		p.buf.writeRune(rune(v))
	default:
		p.buf.writeString(fallback(verb, v))
	}
}

// writeInt handles %d formatting
func writeInt(p *printer, verb rune, arg any) {
	switch v := arg.(type) {
	case int:
		p.buf.writeString(strconv.Itoa(v))
	case int8:
		p.buf.writeString(strconv.Itoa(int(v)))
	case int16:
		p.buf.writeString(strconv.Itoa(int(v)))
	case int32:
		p.buf.writeString(strconv.Itoa(int(v)))
	case int64:
		p.buf.writeString(strconv.Itoa(int(v)))
	case uint:
		p.buf.writeString(strconv.FormatUint(uint64(v), 10))
	case uint8:
		p.buf.writeString(strconv.FormatUint(uint64(v), 10))
	case uint16:
		p.buf.writeString(strconv.FormatUint(uint64(v), 10))
	case uint32:
		p.buf.writeString(strconv.FormatUint(uint64(v), 10))
	case uint64:
		p.buf.writeString(strconv.FormatUint(v, 10))
	default:
		p.buf.writeString(fallback(verb, v))
	}
}

// writeFloatWithPrecision handles floating-point formatting with precision
func writeFloatWithPrecision(p *printer, verb rune, arg any, precision int) {
	switch v := arg.(type) {
	case float64:
		format := byte(verb)
		if format == 'F' {
			format = 'f'
		}
		if precision < 0 {
			switch format {
			case 'e', 'E':
				precision = 2
			default:
				precision = 6
			}
		}
		p.buf = strconv.AppendFloat(p.buf, v, format, precision, 64)
	default:
		p.buf.writeString(fallback(verb, v))
	}
}

// writeBool handles %t formatting
func writeBool(p *printer, verb rune, arg any) {
	switch v := arg.(type) {
	case bool:
		if v {
			p.buf.writeString("true")
		} else {
			p.buf.writeString("false")
		}
	default:
		p.buf.writeString(fallback(verb, v))
	}
}

// writeHex handles %x formatting
func writeHex(p *printer, verb rune, arg any) {
	switch v := arg.(type) {
	case uint8:
		p.buf.writeString(strconv.FormatUint(uint64(v), 16))
	default:
		p.buf.writeString("(unhandled)")
	}
}

// writeQuotedString handles %q formatting
func writeQuotedString(p *printer, verb rune, arg any) {
	switch v := arg.(type) {
	case string:
		p.buf.writeString(strconv.Quote(v))
	default:
		p.buf.writeString("(unhandled)")
	}
}

// writeType handles %T formatting
func writeType(p *printer, arg any) {
	switch arg.(type) {
	case bool:
		p.buf.writeString("bool")
	case int:
		p.buf.writeString("int")
	case int8:
		p.buf.writeString("int8")
	case int16:
		p.buf.writeString("int16")
	case int32:
		p.buf.writeString("int32")
	case int64:
		p.buf.writeString("int64")
	case uint:
		p.buf.writeString("uint")
	case uint8:
		p.buf.writeString("uint8")
	case uint16:
		p.buf.writeString("uint16")
	case uint32:
		p.buf.writeString("uint32")
	case uint64:
		p.buf.writeString("uint64")
	case string:
		p.buf.writeString("string")
	case []byte:
		p.buf.writeString("[]byte")
	case []rune:
		p.buf.writeString("[]rune")
	default:
		p.buf.writeString("unknown")
	}
}

// Fprintf formats according to a format specifier and writes to w.
// Returns the number of bytes written and any write error encountered.
func Fprintf(w io.Writer, format string, a ...any) (n int, err error) {
	p := newPrinter()
	p.doPrintf(format, a)
	return w.Write(p.buf)
}

// Printf formats according to a format specifier and writes to standard output.
// Returns the number of bytes written and any write error encountered.
//
// XXX: Replace with os.Stdout handling when available.
func Printf(format string, a ...any) (n int, err error) {
	var out strings.Builder
	n, err = Fprintf(&out, format, a...)
	print(out.String())
	return n, err
}

// Appendf formats according to a format specifier, appends the result to the byte
// slice, and returns the updated slice.
func Appendf(b []byte, format string, a ...any) []byte {
	p := newPrinter()
	p.doPrintf(format, a)
	return append(b, p.buf...)
}

// Fprint formats using default formats and writes to w.
// Spaces are added between arguments.
// Returns the number of bytes written and any write error encountered.
func Fprint(w io.Writer, a ...any) (n int, err error) {
	p := newPrinter()
	p.doPrint(a)
	return w.Write(p.buf)
}

// Print formats using default formats and writes to standard output.
// Spaces are added between arguments.
// Returns the number of bytes written and any write error encountered.
//
// XXX: Replace with os.Stdout handling when available.
func Print(a ...any) (n int, err error) {
	var out strings.Builder
	n, err = Fprint(&out, a...)
	print(out.String())
	return n, err
}

// Append formats using default formats, appends to b, and returns the updated slice.
// Spaces are added between arguments.
func Append(b []byte, a ...any) []byte {
	p := newPrinter()
	p.doPrint(a)
	return append(b, p.buf...)
}

// Fprintln formats using default formats and writes to w with newline.
// Returns the number of bytes written and any write error encountered.
func Fprintln(w io.Writer, a ...any) (n int, err error) {
	p := newPrinter()
	p.doPrintln(a)
	return w.Write(p.buf)
}

// Println formats using default formats and writes to standard output with newline.
// Returns the number of bytes written and any write error encountered.
//
// XXX: Replace with os.Stdout handling when available.
func Println(a ...any) (n int, err error) {
	var out strings.Builder
	n, err = Fprintln(&out, a...)
	print(out.String())
	return n, err
}

// Sprintln formats using default formats and returns the string with newline.
// Spaces are always added between arguments.
func Sprintln(a ...any) string {
	p := newPrinter()
	p.doPrintln(a)
	return string(p.buf)
}

// Appendln formats using default formats, appends to b, and returns the updated slice.
// Appends a newline after the last argument.
func Appendln(b []byte, a ...any) []byte {
	p := newPrinter()
	p.doPrintln(a)
	return append(b, p.buf...)
}

// This function is used to mimic Go's fmt.Sprintf
// specific behaviour of showing verb/type mismatches,
// where for example:
//
//	fmt.Sprintf("%d", "foo") gives "%!d(string=foo)"
//
// Here:
//
//	fallback("s", 8) -> "%!s(int=8)"
//	fallback("d", nil) -> "%!d(<nil>)", and so on.f
func fallback(verb rune, arg any) string {
	var s string
	switch v := arg.(type) {
	case string:
		s = "string=" + v
	case (interface{ String() string }):
		s = "string=" + v.String()
	case error:
		// note: also "string=" in Go fmt
		s = "string=" + v.Error()
	case float64:
		s = "float64=" + Sprintf("%f", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// note: rune, byte would be dups, being aliases
		if typename, e := typeToString(v); e == nil {
			s = typename + "=" + Sprintf("%d", v)
		} else {
			panic("ufmt: unexpected type error")
		}
	case bool:
		s = "bool=" + strconv.FormatBool(v)
	case nil:
		s = "<nil>"
	default:
		s = "(unhandled)"
	}
	return "%!" + string(verb) + "(" + s + ")"
}

// typeToString returns the name of basic Go types as string.
func typeToString(v any) (string, error) {
	switch v.(type) {
	case string:
		return "string", nil
	case int:
		return "int", nil
	case int8:
		return "int8", nil
	case int16:
		return "int16", nil
	case int32:
		return "int32", nil
	case int64:
		return "int64", nil
	case uint:
		return "uint", nil
	case uint8:
		return "uint8", nil
	case uint16:
		return "uint16", nil
	case uint32:
		return "uint32", nil
	case uint64:
		return "uint64", nil
	case float32:
		return "float32", nil
	case float64:
		return "float64", nil
	case bool:
		return "bool", nil
	default:
		return "", errors.New("unsupported type")
	}
}

// errMsg implements the error interface for formatted error strings.
type errMsg struct {
	msg string
}

// Error returns the formatted error message.
func (e *errMsg) Error() string {
	return e.msg
}

// Errorf formats according to a format specifier and returns an error value.
// Supports the same verbs as Sprintf. See Sprintf documentation for details.
func Errorf(format string, args ...any) error {
	return &errMsg{Sprintf(format, args...)}
}
