// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"errors"
	"io"
	"math"
	"math/bits"
	"strconv"
	"strings"
	"unicode/utf8"
)

// kind represents an encoding type.
type kind uint8

const (
	_ kind = (1 << iota) / 2
	name
	scalar
	objectOpen
	objectClose
	arrayOpen
	arrayClose
)

// Encoder provides methods to write out JSON constructs and values. The user is
// responsible for producing valid sequences of JSON constructs and values.
type Encoder struct {
	AminoSupport bool

	indent   string
	lastKind kind
	indents  []byte

	out io.Writer
}

// NewEncoder returns an Encoder.
//
// If indent is a non-empty string, it causes every entry for an Array or Object
// to be preceded by the indent and trailed by a newline.
func NewEncoder(buf []byte, w io.Writer, indent string) (*Encoder, error) {
	e := &Encoder{
		out: w,
	}
	if len(indent) > 0 {
		if strings.Trim(indent, " \t") != "" {
			return nil, errors.New("indent may only be composed of space or tab characters")
		}
		e.indent = indent
	}
	return e, nil
}

func (e *Encoder) writeRawBytes(b []byte) {
	if _, err := e.out.Write(b); err != nil {
		panic("encoder: unable to write to output: " + err.Error())
	}
}

func (e *Encoder) writeRawString(s string) {
	e.writeRawBytes([]byte(s))
}

func (e *Encoder) writeRawRune(r rune) {
	e.writeRawBytes([]byte{byte(r)})
}

// WriteNull writes out the null value.
func (e *Encoder) WriteNull() {
	e.prepareNext(scalar)
	e.writeRawString("null")
}

// WriteBool writes out the given boolean value.
func (e *Encoder) WriteBool(b bool) {
	e.prepareNext(scalar)
	if b {
		e.writeRawString("true")
	} else {
		e.writeRawString("false")
	}
}

// WriteString writes out the given string in JSON string value. Returns error
// if input string contains invalid UTF-8.
func (e *Encoder) WriteString(s string) error {
	e.prepareNext(scalar)
	var err error
	if err = e.appendString(s); err != nil {
		return err
	}
	return nil
}

// Sentinel error used for indicating invalid UTF-8.
var errInvalidUTF8 = errors.New("invalid UTF-8")

func (e *Encoder) appendString(in string) error {
	e.writeRawRune('"')
	i := indexNeedEscapeInString(in)
	e.writeRawString(in[:i])
	in = in[i:]
	for len(in) > 0 {
		switch r, n := utf8.DecodeRuneInString(in); {
		case r == utf8.RuneError && n == 1:
			return errInvalidUTF8
		case r < ' ' || r == '"' || r == '\\':
			e.writeRawRune('\\')
			switch r {
			case '"', '\\':
				e.writeRawRune(r)
			case '\b':
				e.writeRawRune('b')
			case '\f':
				e.writeRawRune('f')
			case '\n':
				e.writeRawRune('n')
			case '\r':
				e.writeRawRune('r')
			case '\t':
				e.writeRawRune('t')
			default:
				e.writeRawRune('u')
				e.writeRawString("0000"[1+(bits.Len32(uint32(r))-1)/4:])
				e.writeRawString(strconv.FormatUint(uint64(r), 16))
			}
			in = in[n:]
		default:
			i := indexNeedEscapeInString(in[n:])
			e.writeRawString(in[:n+i])
			in = in[n+i:]
		}
	}
	e.writeRawRune('"')
	return nil
}

// indexNeedEscapeInString returns the index of the character that needs
// escaping. If no characters need escaping, this returns the input length.
func indexNeedEscapeInString(s string) int {
	for i, r := range s {
		if r < ' ' || r == '\\' || r == '"' || r == utf8.RuneError {
			return i
		}
	}
	return len(s)
}

// WriteFloat writes out the given float and bitSize in JSON number value.
func (e *Encoder) WriteFloat(n float64, bitSize int) {
	e.prepareNext(scalar)
	e.appendFloat(n, bitSize)
}

// appendFloat formats given float in bitSize, and appends to the given []byte.
func (e *Encoder) appendFloat(n float64, bitSize int) {
	switch {
	case math.IsNaN(n):
		e.writeRawString(`"NaN"`)
		return
	case math.IsInf(n, +1):
		e.writeRawString(`"Infinity"`)
		return
	case math.IsInf(n, -1):
		e.writeRawString(`"-Infinity"`)
		return
	}

	// JSON number formatting logic based on encoding/json.
	// See floatEncoder.encode for reference.
	fmt := byte('f')
	if abs := math.Abs(n); abs != 0 {
		if bitSize == 64 && (abs < 1e-6 || abs >= 1e21) ||
			bitSize == 32 && (float32(abs) < 1e-6 || float32(abs) >= 1e21) {
			fmt = 'e'
		}
	}

	out := strconv.AppendFloat([]byte{}, n, fmt, -1, bitSize)
	if fmt == 'e' {
		n := len(out)
		if n >= 4 && out[n-4] == 'e' && out[n-3] == '-' && out[n-2] == '0' {
			out[n-2] = out[n-1]
			out = out[:n-1]
		}
	}

	e.writeRawBytes(out)
}

// ----------------------------------------
// Signed

// WriteInt writes out the given signed integer in JSON number value.
func (e *Encoder) WriteInt(x int) { e.WriteInt64(int64(x)) }

// WriteInt8 writes out the given signed 8-bit integer in JSON number value.
func (e *Encoder) WriteInt8(x int8) { e.WriteInt64(int64(x)) }

// WriteInt16 writes out the given signed 16-bit integer in JSON number value.
func (e *Encoder) WriteInt16(x int16) { e.WriteInt64(int64(x)) }

// WriteInt32 writes out the given signed 32-bit integer in JSON number value.
func (e *Encoder) WriteInt32(x int32) { e.WriteInt64(int64(x)) }

// WriteInt64 writes out the given signed 64-bit integer in JSON number value.
func (e *Encoder) WriteInt64(x int64) {
	e.prepareNext(scalar)
	e.writeRawString(strconv.FormatInt(x, 10))
}

// ----------------------------------------
// Unsigned

// WriteUint writes out the given unsigned integer in JSON number value.
func (e *Encoder) WriteUint(x uint) { e.WriteUint64(uint64(x)) }

// WriteUint8 writes out the given unsigned 8-bit integer in JSON number value.
func (e *Encoder) WriteUint8(x uint8) { e.WriteUint64(uint64(x)) }

// WriteUint16 writes out the given unsigned 16-bit integer in JSON number value.
func (e *Encoder) WriteUint16(x uint16) { e.WriteUint64(uint64(x)) }

// WriteUint32 writes out the given unsigned 32-bit integer in JSON number value.
func (e *Encoder) WriteUint32(x uint32) { e.WriteUint64(uint64(x)) }

// WriteUint64 writes out the given unsigned 64-bit integer in JSON number value.
func (e *Encoder) WriteUint64(x uint64) {
	e.prepareNext(scalar)
	e.writeRawString(strconv.FormatUint(x, 10))
}

// ----------------------------------------
// Float

// WriteFloat32 writes out the given 32-bit floating point number in JSON number value.
func (e *Encoder) WriteFloat32(x float32) {
	e.prepareNext(scalar)
	e.writeRawString(strconv.FormatFloat(float64(x), 'f', -1, 32))
}

// WriteFloat64 writes out the given 64-bit floating point number in JSON number value.
func (e *Encoder) WriteFloat64(x float64) {
	e.prepareNext(scalar)
	e.writeRawString(strconv.FormatFloat(float64(x), 'f', -1, 64))
}

// StartObject writes out the '{' symbol.
func (e *Encoder) StartObject() {
	e.prepareNext(objectOpen)
	e.writeRawRune('{')
}

// EndObject writes out the '}' symbol.
func (e *Encoder) EndObject() {
	e.prepareNext(objectClose)
	e.writeRawRune('}')
}

// WriteName writes out the given string in JSON string value and the name
// separator ':'. Returns error if input string contains invalid UTF-8.
func (e *Encoder) WriteName(s string) error {
	e.prepareNext(name)
	// Append to output regardless of error.
	if err := e.appendString(s); err != nil {
		return err
	}

	e.writeRawRune(':')
	return nil
}

// StartArray writes out the '[' symbol.
func (e *Encoder) StartArray() {
	e.prepareNext(arrayOpen)
	e.writeRawRune('[')
}

// EndArray writes out the ']' symbol.
func (e *Encoder) EndArray() {
	e.prepareNext(arrayClose)
	e.writeRawRune(']')
}

// prepareNext adds possible comma and indentation for the next value based
// on last type and indent option. It also updates lastKind to next.
func (e *Encoder) prepareNext(next kind) {
	defer func() {
		// Set lastKind to next.
		e.lastKind = next
	}()

	if len(e.indent) == 0 {
		// Need to add comma on the following condition.
		if e.lastKind&(scalar|objectClose|arrayClose) != 0 &&
			next&(name|scalar|objectOpen|arrayOpen) != 0 {
			e.writeRawRune(',')
		}
		return
	}

	switch {
	case e.lastKind&(objectOpen|arrayOpen) != 0:
		// If next type is NOT closing, add indent and newline.
		if next&(objectClose|arrayClose) == 0 {
			e.indents = append(e.indents, e.indent...)
			e.writeRawRune('\n')
			e.writeRawBytes(e.indents)
		}

	case e.lastKind&(scalar|objectClose|arrayClose) != 0:
		switch {
		// If next type is either a value or name, add comma and newline.
		case next&(name|scalar|objectOpen|arrayOpen) != 0:
			e.writeRawString(",\n")

		// If next type is a closing object or array, adjust indentation.
		case next&(objectClose|arrayClose) != 0:
			e.indents = e.indents[:len(e.indents)-len(e.indent)]
			e.writeRawRune('\n')
		}
		e.writeRawBytes(e.indents)

	case e.lastKind&name != 0:
		e.writeRawRune(' ')
	}
}
