// Package ufmt provides utility functions for formatting strings, similarly
// to the Go package "fmt", of which only a subset is currently supported
// (hence the name µfmt - micro fmt).
package ufmt

import (
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

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

type printer struct {
	buf buffer
}

func newPrinter() *printer {
	return &printer{}
}

// Println formats using the default formats for its operands and writes to standard output.
// Println writes the given arguments to standard output with spaces between arguments
// and a newline at the end.
func (p *printer) doPrint(args []any) {
	for i, arg := range args {
		// Add a space between two arguments.
		if i > 0 {
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

// doPrintln is like doPrint but add a newline after the last argument.
func (p *printer) doPrintln(a []any) {
	p.doPrint(a)
	p.buf.writeByte('\n')
}

// Sprintf offers similar functionality to Go's fmt.Sprintf, or the sprintf
// equivalent available in many languages, including C/C++.
// The number of args passed must exactly match the arguments consumed by the format.
// A limited number of formatting verbs and features are currently supported,
// hence the name ufmt (µfmt, micro-fmt).
//
// The currently formatted verbs are the following:
//
//	%s: places a string value directly.
//	    If the value implements the interface interface{ String() string },
//	    the String() method is called to retrieve the value. Same about Error()
//	    string.
//	%c: formats the character represented by Unicode code point
//
//	%d: formats an integer value using package "strconv".
//	    Currently supports only uint, uint64, int, int64.
//	%f: formats a float value, with a default precision of 6.
//	%e: formats a float with scientific notation; 1.23456e+78
//	%E: formats a float with scientific notation; 1.23456E+78
//	%F: The same as %f
//	%g: formats a float value with %e for large exponents, and %f with full precision for smaller numbers
//	%G: formats a float value with %G for large exponents, and %F with full precision for smaller numbers
//	%t: formats a boolean value to "true" or "false".
//	%x: formats an integer value as a hexadecimal string.
//	    Currently supports only uint8, []uint8, [32]uint8.
//	%c: formats a rune value as a string.
//	    Currently supports only rune, int.
//	%q: formats a string value as a quoted string.
//	%T: formats the type of the value.
//	%v: formats the value with a default representation appropriate for the value's type
//	%%: outputs a literal %. Does not consume an argument.
func (p *printer) doPrintf(format string, args []any) {
	// we use runes to handle multi-byte characters
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

		verb := sTor[i+1]
		if verb == '%' {
			p.buf.writeRune('%')
			i += 2
			continue
		}

		if argNum > argLen {
			panic("invalid number of arguments to ufmt.Sprintf")
		}
		arg := args[argNum]
		argNum++

		switch verb {
		case 'v':
			switch v := arg.(type) {
			case nil:
				p.buf.writeString("<nil>")
			case bool:
				if v {
					p.buf.writeString("true")
				} else {
					p.buf.writeString("false")
				}
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
			case interface{ String() string }:
				p.buf.writeString(v.String())
			case error:
				p.buf.writeString(v.Error())
			default:
				p.buf.writeString(fallback(verb, v))
			}
		case 's':
			switch v := arg.(type) {
			case interface{ String() string }:
				p.buf.writeString(v.String())
			case error:
				p.buf.writeString(v.Error())
			case string:
				p.buf.writeString(v)
			default:
				p.buf.writeString(fallback(verb, v))
			}
		case 'c':
			switch v := arg.(type) {
			// rune is int32. Exclude overflowing numeric types and dups (byte, int32):
			case rune:
				p.buf.writeString(string(v))
			case int:
				p.buf.writeString(string(v))
			case int8:
				p.buf.writeString(string(v))
			case int16:
				p.buf.writeString(string(v))
			case uint:
				p.buf.writeString(string(v))
			case uint8:
				p.buf.writeString(string(v))
			case uint16:
				p.buf.writeString(string(v))
			default:
				p.buf.writeString(fallback(verb, v))
			}
		case 'd':
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
		case 'e', 'E', 'f', 'F', 'g', 'G':
			switch v := arg.(type) {
			case float64:
				switch verb {
				case 'e':
					p.buf.writeString(strconv.FormatFloat(v, byte('e'), -1, 64))
				case 'E':
					p.buf.writeString(strconv.FormatFloat(v, byte('E'), -1, 64))
				case 'f', 'F':
					p.buf.writeString(strconv.FormatFloat(v, byte('f'), 6, 64))
				case 'g':
					p.buf.writeString(strconv.FormatFloat(v, byte('g'), -1, 64))
				case 'G':
					p.buf.writeString(strconv.FormatFloat(v, byte('G'), -1, 64))
				}
			default:
				p.buf.writeString(fallback(verb, v))
			}
		case 't':
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
		case 'x':
			switch v := arg.(type) {
			case uint8:
				p.buf.writeString(strconv.FormatUint(uint64(v), 16))
			default:
				p.buf.writeString("(unhandled)")
			}
		case 'q':
			switch v := arg.(type) {
			case string:
				p.buf.writeString(strconv.Quote(v))
			default:
				p.buf.writeString("(unhandled)")
			}
		case 'T':
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
		// % handled before, as it does not consume an argument
		default:
			p.buf.writeString("(unhandled verb: %" + string(verb) + ")")
		}

		i += 2
	}
	if argNum < argLen {
		panic("too many arguments to ufmt.Sprintf")
	}
}

// These routines end in 'f' and take a format string.

// Fprintf formats according to a format specifier and writes to w.
// It returns the number of bytes written and any write error encountered.
func Fprintf(w io.Writer, format string, a ...any) (n int, err error) {
	p := newPrinter()
	p.doPrintf(format, a)
	return w.Write(p.buf)
}

// Printf formats according to a format specifier and writes to standard output.
// It returns the number of bytes written and any write error encountered.
func Printf(format string, a ...any) (n int, err error) {
	// XXX: update this after gno supports os.Stdout
	var out strings.Builder
	n, err = Fprintf(&out, format, a)
	print(out.String())
	return n, err
}

// Sprintf formats according to a format specifier and returns the resulting string.
func Sprintf(format string, a ...any) string {
	p := newPrinter()
	p.doPrintf(format, a)
	return string(p.buf)
}

// Appendf formats according to a format specifier, appends the result to the byte
// slice, and returns the updated slice.
func Appendf(b []byte, format string, a ...any) []byte {
	p := newPrinter()
	p.doPrintf(format, a)
	return append(b, p.buf...)
}

// Fprint formats using the default formats for its operands and writes to w.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func Fprint(w io.Writer, a ...any) (n int, err error) {
	p := newPrinter()
	p.doPrint(a)
	return w.Write(p.buf)
}

// Print formats using the default formats for its operands and writes to standard output.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func Print(a ...any) (n int, err error) {
	// XXX: update this after gno supports os.Stdout
	var out strings.Builder
	n, err = Fprint(&out, a...)
	print(out.String())
	return n, err
}

// Sprint formats using the default formats for its operands and returns the resulting string.
// Spaces are added between operands when neither is a string.
func Sprint(a ...any) string {
	p := newPrinter()
	p.doPrint(a)
	s := string(p.buf)
	return s
}

// Append formats using the default formats for its operands, appends the result to
// the byte slice, and returns the updated slice.
func Append(b []byte, a ...any) []byte {
	p := newPrinter()
	p.doPrint(a)
	return append(b, p.buf...)
}

// These routines end in 'ln', do not take a format string,
// always add spaces between operands, and add a newline
// after the last operand.

// Fprintln formats using the default formats for its operands and writes to w.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func Fprintln(w io.Writer, a ...any) (n int, err error) {
	p := newPrinter()
	p.doPrintln(a)
	return w.Write(p.buf)
}

// Println formats using the default formats for its operands and writes to standard output.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func Println(a ...any) (n int, err error) {
	// XXX: update this after gno supports os.Stdout
	var out strings.Builder
	n, err = Fprintln(&out, a...)
	if err != nil {
		print(out.String())
	}

	return n, err
}

// Sprintln formats using the default formats for its operands and returns the resulting string.
// Spaces are always added between operands and a newline is appended.
func Sprintln(a ...any) string {
	p := newPrinter()
	p.doPrintln(a)
	return string(p.buf)

}

// Appendln formats using the default formats for its operands, appends the result
// to the byte slice, and returns the updated slice. Spaces are always added
// between operands and a newline is appended.
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
//	fallback("d", nil) -> "%!d(<nil>)", and so on.
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
		if typename, e := typeToString(v); e != nil {
			panic("should not happen")
		} else {
			s = typename + "=" + Sprintf("%d", v)
		}
	case bool:
		if v {
			s = "bool=true"
		} else {
			s = "bool=false"
		}
	case nil:
		s = "<nil>"
	default:
		s = "(unhandled)"
	}
	return "%!" + string(verb) + "(" + s + ")"
}

// Get the name of the type of `v` as a string.
// The recognized type of v is currently limited to native non-composite types.
// An error is returned otherwise.
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
		return "", errors.New("(unsupported type)")
	}
}

// errMsg implements the error interface.
type errMsg struct {
	msg string
}

// Error defines the requirements of the error interface.
// It functions similarly to Go's errors.New()
func (e *errMsg) Error() string {
	return e.msg
}

// Errorf is a function that mirrors the functionality of fmt.Errorf.
//
// It takes a format string and arguments to create a formatted string,
// then sets this string as the 'msg' field of an errMsg struct and returns a pointer to this struct.
//
// This function operates in a similar manner to Go's fmt.Errorf,
// providing a way to create formatted error messages.
//
// The currently formatted verbs are the following:
//
//	%s: places a string value directly.
//	    If the value implements the interface interface{ String() string },
//	    the String() method is called to retrieve the value. Same for error.
//	%c: formats the character represented by Unicode code point
//	%d: formats an integer value using package "strconv".
//	    Currently supports only uint, uint64, int, int64.
//	%t: formats a boolean value to "true" or "false".
//	%%: outputs a literal %. Does not consume an argument.
func Errorf(format string, args ...any) error {
	return &errMsg{Sprintf(format, args...)}
}
