package json

import (
	"bytes"
	"errors"
	"math"
	"strconv"
)

const (
	absMinInt64 = 1 << 63
	maxInt64    = 1<<63 - 1
	maxUint64   = 1<<64 - 1
	intSize     = 32 << (^uint(0) >> 63)
	IntSize     = intSize
)

func ParseString(data []byte) (string, error) {
	var buf [UnescapeStackBufSize]byte

	bf, err := Unescape(data, buf[:])
	if err != nil {
		return "", MalformedString
	}

	return string(bf), nil
}

func ParseBool(data []byte) (bool, error) {
	switch {
	case bytes.Equal(data, []byte("true")):
		return true, nil
	case bytes.Equal(data, []byte("false")):
		return false, nil
	default:
		return false, MalformedValue
	}
}

func ParseFloat(bytes []byte) (value float64, err error) {
	if len(bytes) == 0 {
		return -1, EmptyBytes
	}

	neg, bytes := trimNegativeSign(bytes)

	var exponentPart []byte
	for i, c := range bytes {
		if c == 'e' || c == 'E' {
			exponentPart = bytes[i+1:]
			bytes = bytes[:i]
			break
		}
	}

	man, exp10, err := extractMantissaAndExp10(bytes)
	if err != nil {
		return -1, err
	}

	if len(exponentPart) > 0 {
		exp, err := strconv.Atoi(string(exponentPart))
		if err != nil {
			return -1, InvalidExponents
		}
		exp10 += exp
	}

	f, success := eiselLemire64(man, exp10, neg)
	if !success {
		return 0, nil
	}

	return f, nil
}

func ParseInt(bytes []byte) (v int64, err error) {
	if len(bytes) == 0 {
		return 0, EmptyBytes
	}

	neg, bytes := trimNegativeSign(bytes)

	var n uint64 = 0
	for _, c := range bytes {
		if notDigit(c) {
			return 0, NonDigitCharacters
		}

		if n > maxUint64/10 {
			return 0, Overflow
		}

		n *= 10

		n1 := n + uint64(c-'0')
		if n1 < n {
			return 0, Overflow
		}

		n = n1
	}

	if n > maxInt64 {
		if neg && n == absMinInt64 {
			return -absMinInt64, nil
		}

		return 0, Overflow
	}

	if neg {
		return -int64(n), nil
	}

	return int64(n), nil
}

func ParseUint(b []byte, base, bitSize int) (v uint64, err error) {
	if len(b) == 0 {
		return 0, EmptyBytes
	}

	base0 := base == 0
	s0 := string(b)

	switch {
	case 2 <= base && base <= 36:
		// do nothing
	case base == 0:
		base = 10
		if b[0] == '0' {
			switch {
			case len(b) >= 3 && lower(b[1]) == 'b':
				base = 2
				b = b[2:]
			case len(b) >= 3 && lower(b[1]) == 'o':
				base = 8
				b = b[2:]
			case len(b) >= 3 && lower(b[1]) == 'x':
				base = 16
				b = b[2:]
			default:
				base = 8
				b = b[1:]
			}
		}
	default:
		return 0, errors.New("Base Error")
	}

	if bitSize == 0 {
		bitSize = IntSize
	} else if bitSize < 0 || bitSize > 64 {
		return 0, errors.New("BitSize Error")
	}

	var cutoff uint64
	switch base {
	case 10:
		cutoff = maxUint64/10 + 1
	case 16:
		cutoff = maxUint64/16 + 1
	default:
		cutoff = maxUint64/uint64(base) + 1
	}

	maxVal := uint64(1)<<uint(bitSize) - 1
	underscores := false
	var n uint64
	for _, c := range b {
		var d byte
		switch {
		case c == '_' && base0:
			underscores = true
			continue
		case '0' <= c && c <= '9':
			d = c - '0'
		case 'a' <= lower(c) && lower(c) <= 'z':
			d = lower(c) - 'a' + 10
		default:
			return 0, errors.New("ParseUint Syntax Error")
		}

		if d >= byte(base) {
			return 0, errors.New("ParseUint Syntax Error")
		}

		if n >= cutoff {
			return maxVal, errors.New("ParseUint Range Error")
		}
		n *= uint64(base)

		n1 := n + uint64(d)
		if n1 < n || n1 > maxVal {
			return maxVal, errors.New("ParseUint Range Error")
		}
		n = n1
	}

	if underscores && !underscoreOK(s0) {
		return 0, errors.New("ParseUint Syntax Error")
	}

	return n, nil
}

// lower is a lower-case letter if and only if
// c is either that lower-case letter or the equivalent upper-case letter.
// Instead of writing c == 'x' || c == 'X' one can write lower(c) == 'x'.
// Note that lower of non-letters can produce other non-letters.
func lower(c byte) byte {
	return c | ('x' - 'X')
}

func extractMantissaAndExp10(bytes []byte) (uint64, int, error) {
	var man uint64
	var exp10 int
	decimalFound := false

	for _, c := range bytes {
		if c == '.' {
			if decimalFound {
				return 0, 0, MultipleDecimalPoints
			}
			decimalFound = true
			continue
		}

		if notDigit(c) {
			return 0, 0, NonDigitCharacters
		}

		digit := uint64(c - '0')

		if man > (maxUint64-digit)/10 {
			return 0, 0, Overflow
		}

		man = man*10 + digit

		if decimalFound {
			exp10--
		}
	}

	return man, exp10, nil
}

func extractValueTypeFromToken(b byte, value []byte) (dataType ValueType, offset int, err error) {
	switch b {
	case 't', 'f':
		if bytes.Equal(value, TrueLiteral) || bytes.Equal(value, FalseLiteral) {
			dataType = Boolean
		}
		return Unknown, offset, UnknownValueType
	case 'u', 'n':
		if bytes.Equal(value, NullLiteral) {
			dataType = Null
		}
		return Unknown, offset, UnknownValueType
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		dataType = Number
	default:
		return Unknown, offset, UnknownValueType
	}
}

// underscoreOK reports whether the underscores in s are allowed.
// Checking them in this one function lets all the parsers skip over them simply.
// Underscore must appear only between digits or between a base prefix and a digit.
func underscoreOK(s string) bool {
	// saw tracks the last character (class) we saw:
	// ^ for beginning of number,
	// 0 for a digit or base prefix,
	// _ for an underscore,
	// ! for none of the above.
	saw := '^'
	i := 0

	// Optional sign.
	if len(s) >= 1 && (s[0] == '-' || s[0] == '+') {
		s = s[1:]
	}

	// Optional base prefix.
	hex := false
	if len(s) >= 2 && s[0] == '0' && (lower(s[1]) == 'b' || lower(s[1]) == 'o' || lower(s[1]) == 'x') {
		i = 2
		saw = '0' // base prefix counts as a digit for "underscore as digit separator"
		hex = lower(s[1]) == 'x'
	}

	// Number proper.
	for ; i < len(s); i++ {
		// Digits are always okay.
		if '0' <= s[i] && s[i] <= '9' || hex && 'a' <= lower(s[i]) && lower(s[i]) <= 'f' {
			saw = '0'
			continue
		}
		// Underscore must follow digit.
		if s[i] == '_' {
			if saw != '0' {
				return false
			}
			saw = '_'
			continue
		}
		// Underscore must also be followed by digit.
		if saw == '_' {
			return false
		}
		// Saw non-digit, non-underscore.
		saw = '!'
	}

	return saw != '_'
}

var (
	trueLiteral  = []byte("true")
	falseLiteral = []byte("false")
	nullLiteral  = []byte("null")
)

func getType(data []byte, offset int) ([]byte, ValueType, int, error) {
	if len(data) == 0 {
		return nil, Unknown, offset, errors.New("no JSON data provided")
	}

	dataType, endOffset, err := parseValue(data, offset)
	if err != nil {
		return nil, dataType, offset, err
	}

	return data[offset:endOffset], dataType, endOffset, nil
}

func parseValue(data []byte, offset int) (ValueType, int, error) {
	switch data[offset] {
	case '"':
		return parseString(data, offset)
	case '[', '{':
		return parseContainer(data, offset)
	}

	return parsePrimitive(data, offset)
}

// parseString parses a JSON string and returns its type and the end position.
func parseString(data []byte, offset int) (ValueType, int, error) {
	if idx, _ := stringEnd(data[offset+1:]); idx != -1 {
		return String, offset + idx + 1, nil
	}
	return String, offset, errors.New("malformed string")
}

// parseContainer parses a JSON array or object and returns its type and the end position.
func parseContainer(data []byte, offset int) (ValueType, int, error) {
	var containerType ValueType
	var closing byte

	if data[offset] == '[' {
		containerType = Array
		closing = ']'
	} else {
		containerType = Object
		closing = '}'
	}

	endOffset := blockEnd(data[offset:], data[offset], closing)
	if endOffset == -1 {
		return containerType, offset, MalformedType
	}

	return containerType, offset + endOffset, nil
}

func parsePrimitive(data []byte, offset int) (ValueType, int, error) {
	end := tokenEnd(data[offset:])
	if end == -1 {
		return Unknown, offset, errors.New("malformed value")
	}
	value := data[offset : offset+end]

	switch data[offset] {
	case 't', 'f':
		_boolValue, err := ParseBool(value)
		if err != nil {
			return Unknown, offset, err
		}
		return Boolean, offset + end, nil
	case 'n':
		if bytes.Equal(value, nullLiteral) {
			return Null, offset + end, nil
		}
		return Unknown, offset, errors.New("unknown null type")
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		if !bytes.ContainsAny(value, ".eE") {
			_intValue, err := ParseInt(value)
			if err != nil {
				return Number, offset + end, err
			}
		}

		if _floatValue, err := ParseFloat(value); err != nil {
			return Number, offset + end, err
		}

		return Number, offset + end, nil
	}

	return Unknown, offset, errors.New("unknown value type")
}

// ArrayEach is used when iterating arrays, accepts a callback function with the same return arguments as `Get`.
func ArrayEach(data []byte, f func(value []byte, dataType ValueType, offset int, err error), keys ...string) (offset int, err error) {
	if len(data) == 0 {
		return -1, MalformedObject
	}

	nT, err := nextToken(data)
	if err != nil {
		return -1, MalformedJson
	}

	offset = nT + 1

	if len(keys) > 0 {
		if offset, _ = searchKeys(data, keys...); offset == -1 {
			return offset, KeyPathNotFoundError
		}

		// Go to closest value
		nO, err := nextToken(data[offset:])
		if err != nil {
			return offset, MalformedJson
		}

		offset += nO

		if data[offset] != '[' {
			return offset, MalformedArray
		}

		offset++
	}

	nO, err := nextToken(data[offset:])
	if err != nil {
		return offset, MalformedJson
	}

	offset += nO

	if data[offset] == ']' {
		return offset, nil
	}

	for true {
		v, t, o, e := Get(data[offset:])

		if e != nil {
			return offset, e
		}

		if o == 0 {
			break
		}

		if t != NotExist {
			f(v, t, offset+o-len(v), e)
		}

		if e != nil {
			break
		}

		offset += o

		skipToToken, err := nextToken(data[offset:])
		if err != nil {
			return offset, MalformedArray
		}
		offset += skipToToken

		if data[offset] == ']' {
			break
		}

		if data[offset] != ',' {
			return offset, MalformedArray
		}

		offset++
	}

	return offset, nil
}

func internalGet(data []byte, keys ...string) (value []byte, dataType ValueType, offset, endOffset int, err error) {
	if len(keys) > 0 {
		offset, err = searchKeys(data, keys...)
		if err != nil {
			return nil, NotExist, -1, -1, err
		}
	}

	// Go to closest value
	nO, err := nextToken(data[offset:])
	if err != nil {
		return nil, NotExist, offset, -1, MalformedJson
	}

	offset += nO
	value, dataType, endOffset, err = getType(data, offset)
	if err != nil {
		return value, dataType, offset, endOffset, err
	}

	// Strip quotes from string values
	if dataType == String {
		value = value[1 : len(value)-1]
	}

	return value[:len(value):len(value)], dataType, offset, endOffset, nil
}
