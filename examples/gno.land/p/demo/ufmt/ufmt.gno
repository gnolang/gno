// Package ufmt provides utility functions for formatting strings, similarly to
// the Go package "fmt", of which only a subset is currently supported (hence
// the name µfmt - micro fmt). It includes functions like Printf, Sprintf,
// Fprintf, and Errorf.
// Supported formatting verbs are documented in the Sprintf function.
package ufmt

import (
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

// the different flags
const (
	flagHashtag uint8 = 1
	flagZero    uint8 = 1 << 1
	flagMinus   uint8 = 1 << 2
	flagSpace   uint8 = 1 << 3
	flagPlus    uint8 = 1 << 4
)

// the different len modifiers
const (
	lenNo uint8 = iota
	lenHH
	lenH
	lenL
	lenLL
	lenZ
)

// printerMeta holds formatting metadata.
type printerMeta struct {
	flags     uint8
	padding   uint
	length    uint8
	width     uint
	precision int
}

// printer holds state for formatting operations.
type printer struct {
	buf  buffer
	meta printerMeta
}

// parseDigit parses a digit rune and returns its integer value.
// Returns -1 if the rune is not a digit.
func parseDigit(r rune) int {
	if '0' <= r && r <= '9' {
		return int(r - '0')
	}
	return -1
}

// parseNumber parses a number from runes starting at the given index.
// Returns the parsed number, the number of runes consumed, and a boolean
// indicating if a number was successfully parsed.
func parseNumber(runes []rune, start int, end int) (int, int, bool) {
	if start >= end {
		return 0, 0, false
	}

	num := 0
	i := 0
	for start+i < end {
		digit := parseDigit(runes[start+i])
		if digit < 0 {
			break
		}
		num = num*10 + digit
		i++
	}

	if i == 0 {
		return 0, 0, false
	}

	return num, i, true
}

// flagMap maps flag characters to their corresponding flag constants.
var flagMap = [256]uint8{
	'-': flagMinus,
	'+': flagPlus,
	' ': flagSpace,
	'0': flagZero,
}

// parseMeta parses formatting metadata from format string runes.
func parseMeta(p *printer, runes []rune, index *int, end int) {
	// Skip the '%' character that starts format specifiers
	*index++
	if *index >= end {
		return
	}

	// Single-pass parsing with lookahead for efficient processing
	for *index < end {
		c := runes[*index]

		switch c {
		case '#':
			// Handle hashtag flag for alternate format
			p.meta.flags |= flagHashtag
			*index++

		case '-', '+', ' ', '0':
			// Process format flags using direct array indexing
			p.meta.flags |= flagMap[c]
			*index++

		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			// Parse width specification (padding)
			if num, consumed, ok := parseNumber(runes, *index, end); ok {
				p.meta.padding = uint(num)
				*index += consumed
			}

		case '.':
			// Parse precision specification (after decimal point)
			*index++
			if *index < end {
				if num, consumed, ok := parseNumber(runes, *index, end); ok {
					p.meta.precision = num
					*index += consumed
				} else {
					// Handle case where only '.' is provided (precision = 0)
					p.meta.precision = 0
				}
			}

		case 'h', 'l', 'z':
			// Use lookahead to check for double-character length modifiers
			// like 'hh' (char) or 'll' (long long) without iterating through options
			if *index+1 < end {
				next := runes[*index+1]
				if c == 'h' && next == 'h' {
					p.meta.length = lenHH
					*index += 2 // Skip both characters
					continue
				} else if c == 'l' && next == 'l' {
					p.meta.length = lenLL
					*index += 2 // Skip both characters
					continue
				}
			}

			// Handle single-character length modifiers
			switch c {
			case 'h':
				p.meta.length = lenH
			case 'l':
				p.meta.length = lenL
			case 'z':
				p.meta.length = lenZ
			}
			*index++

		default:
			// Found a non-metadata character (likely a verb), exit parsing
			goto endParse
		}
	}

endParse:
	// Adjust index to position before the verb for subsequent processing
	if *index < end {
		*index-- // Back up one position to position at the verb
	}
}

// newPrinter creates a new printer with preallocated buffer.
func newPrinter() *printer {
	return &printer{
		buf: make(buffer, 0, 64), // Preallocate buffer with reasonable capacity
		meta: printerMeta{
			precision: 6,
		},
	}
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
			p.buf.writeByte(' ')
		}

		switch v := arg.(type) {
		case string:
			p.buf.writeString(v)
		case (interface{ String() string }):
			p.buf.writeString(v.String())
		case error:
			p.buf.writeString(v.Error())
		case float64:
			p.buf.writeString(strconv.FormatFloat(v, 'f', 6, 64))
		case int:
			p.buf.writeString(strconv.Itoa(v))
		case int8:
			p.buf.writeString(strconv.Itoa(int(v)))
		case int16:
			p.buf.writeString(strconv.Itoa(int(v)))
		case int32:
			p.buf.writeString(strconv.Itoa(int(v)))
		case int64:
			p.buf.writeString(strconv.FormatInt(v, 10))
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
//	    Currently supports only int, uint, uint8, 16, 32, 64 and []uint8.
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
	// TODO: This line consumes memory proportional to the length of the string.
	// If we only need to process UTF-8 encoding, we can reduce memory allocation
	// by iterating through the string directly.
	runes := []rune(format)
	end := len(runes)
	argNum := 0
	argLen := len(args)

	for i := 0; i < end; {
		c := runes[i]

		if c != '%' || i == end-1 {
			p.buf.writeRune(c)
			i++
			continue
		}

		// Reset meta for each format directive
		p.meta = printerMeta{precision: 6}

		parseMeta(p, runes, &i, end)
		i++

		if i >= end {
			break
		}

		verb := runes[i]
		if verb == '%' {
			p.buf.writeByte('%')
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
			writeString(p, verb, arg)
		case 'c':
			writeChar(p, verb, arg)
		case 'd', 'i':
			writeInt(p, verb, arg)
		case 'e', 'E', 'f', 'F', 'g', 'G':
			writeFloat(p, verb, arg)
		case 't':
			writeBool(p, verb, arg)
		case 'x', 'X':
			writeHex(p, arg, verb == 'X')
		case 'q':
			writeQuotedString(p, arg)
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

// applyPadding applies padding according to flags.
func applyPadding(p *printer, content string) string {
	paddingCount := int(p.meta.padding) - len(content)
	if paddingCount <= 0 {
		return content
	}

	// Use a single conditional rather than multiple if checks
	if p.meta.flags&flagZero != 0 {
		return strings.Repeat("0", paddingCount) + content
	}
	if p.meta.flags&flagSpace != 0 || p.meta.flags&flagMinus == 0 {
		return strings.Repeat(" ", paddingCount) + content
	}
	return content
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

// writeString handles %s formatting
func writeString(p *printer, verb rune, arg any) {
	switch v := arg.(type) {
	case (interface{ String() string }):
		p.buf.writeString(v.String())
	case error:
		p.buf.writeString(v.Error())
	case string:
		p.buf.writeString(v)
	default:
		p.buf.writeString(fallback(verb, v))
	}
}

// writeChar handles %c formatting
func writeChar(p *printer, verb rune, arg any) {
	switch v := arg.(type) {
	// rune is int32. Exclude overflowing numeric types and dups (byte, int32):
	case rune:
		p.buf.writeRune(v)
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

// writeInt handles %d & %i formatting
func writeInt(p *printer, verb rune, arg any) {
	var subBuf string
	switch v := arg.(type) {
	case int:
		subBuf = strconv.Itoa(v)
	case int8:
		subBuf = strconv.Itoa(int(v))
	case int16:
		subBuf = strconv.Itoa(int(v))
	case int32:
		subBuf = strconv.Itoa(int(v))
	case int64:
		subBuf = strconv.FormatInt(v, 10)
	case uint:
		subBuf = strconv.FormatUint(uint64(v), 10)
	case uint8:
		subBuf = strconv.FormatUint(uint64(v), 10)
	case uint16:
		subBuf = strconv.FormatUint(uint64(v), 10)
	case uint32:
		subBuf = strconv.FormatUint(uint64(v), 10)
	case uint64:
		subBuf = strconv.FormatUint(v, 10)
	default:
		p.buf.writeString(fallback(verb, v))
		return
	}

	p.buf.writeString(applyPadding(p, subBuf))
}

// writeFloat handles floating-point formatting verbs
func writeFloat(p *printer, verb rune, arg any) {
	var n float64
	var bits int

	switch v := arg.(type) {
	case float64:
		n = v
		bits = 64
	case float32:
		n = float64(v)
		bits = 32
	default:
		p.buf.writeString(fallback(verb, v))
		return
	}

	prec := -1
	verbChar := byte(verb)
	if verbChar == 'f' || verbChar == 'F' {
		prec = p.meta.precision
		verbChar = 'f'
	}

	p.buf.writeString(strconv.FormatFloat(n, verbChar, prec, bits))
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

// writeHex handles %x & %X formatting
func writeHex(p *printer, arg any, big bool) {
	// to avoid the use of intermediate buffers (i.e. strings.Builder)
	var buf buffer
	buf = make(buffer, 0, 32)

	switch v := arg.(type) {
	case int:
		buf.writeString(strconv.FormatUint(uint64(v), 16))
	case uint:
		buf.writeString(strconv.FormatUint(uint64(v), 16))
	case uint8:
		buf.writeString(strconv.FormatUint(uint64(v), 16))
	case uint16:
		buf.writeString(strconv.FormatUint(uint64(v), 16))
	case uint32:
		buf.writeString(strconv.FormatUint(uint64(v), 16))
	case uint64:
		buf.writeString(strconv.FormatUint(v, 16))
	case []uint8:
		for _, b := range v {
			x := strconv.FormatUint(uint64(b), 16)
			if len(x) == 1 {
				buf.writeByte('0')
			}
			buf.writeString(x)
		}
	default:
		p.buf.writeString("(unhandled)")
		return
	}

	result := string(buf)

	// apply padding if needed
	if p.meta.padding > 0 {
		result = applyPadding(p, result)
	}

	// apply 0x prefix if flagHashtag is set
	if p.meta.flags&flagHashtag != 0 {
		result = "0x" + result
	}

	// convert to uppercase if requested
	if big {
		result = strings.ToUpper(result)
	}

	p.buf.writeString(result)
}

// writeQuotedString handles %q formatting
func writeQuotedString(p *printer, arg any) {
	switch v := arg.(type) {
	case string:
		p.buf.writeString(strconv.Quote(v))
	default:
		p.buf.writeString("(unhandled)")
	}
}

// writeType handles %T formatting
func writeType(p *printer, arg any) {
	var typeName string

	switch arg.(type) {
	case bool:
		typeName = "bool"
	case int:
		typeName = "int"
	case int8:
		typeName = "int8"
	case int16:
		typeName = "int16"
	case int32:
		typeName = "int32"
	case int64:
		typeName = "int64"
	case uint:
		typeName = "uint"
	case uint8:
		typeName = "uint8"
	case uint16:
		typeName = "uint16"
	case uint32:
		typeName = "uint32"
	case uint64:
		typeName = "uint64"
	case string:
		typeName = "string"
	case []byte:
		typeName = "[]byte"
	case []rune:
		typeName = "[]rune"
	default:
		typeName = "unknown"
	}

	p.buf.writeString(typeName)
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
	out.Grow(64)
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
	out.Grow(64)
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
	out.Grow(64)
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
		s = "string=" + v.Error()
	case float64:
		s = "float64=" + strconv.FormatFloat(v, 'f', 6, 64)
	case int:
		s = "int=" + strconv.Itoa(v)
	case int8:
		s = "int8=" + strconv.Itoa(int(v))
	case int16:
		s = "int16=" + strconv.Itoa(int(v))
	case int32:
		s = "int32=" + strconv.Itoa(int(v))
	case int64:
		s = "int64=" + strconv.FormatInt(v, 10)
	case uint:
		s = "uint=" + strconv.FormatUint(uint64(v), 10)
	case uint8:
		s = "uint8=" + strconv.FormatUint(uint64(v), 10)
	case uint16:
		s = "uint16=" + strconv.FormatUint(uint64(v), 10)
	case uint32:
		s = "uint32=" + strconv.FormatUint(uint64(v), 10)
	case uint64:
		s = "uint64=" + strconv.FormatUint(v, 10)
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
