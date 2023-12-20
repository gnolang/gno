package json

import (
	"bytes"
	"errors"
	"math"
	"strconv"
)

func ParseStringLiteral(data []byte) (string, error) {
	var buf [UnescapeStackBufSize]byte

	bf, err := Unescape(data, buf[:])
	if err != nil {
		return "", MalformedString
	}

	return string(bf), nil
}

func ParseBoolLiteral(data []byte) (bool, error) {
	switch {
	case bytes.Equal(data, []byte("true")):
		return true, nil
	case bytes.Equal(data, []byte("false")):
		return false, nil
	default:
		return false, MalformedValue
	}
}

func ParseFloatLiteral(bytes []byte) (value float64, err error) {
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

func ParseIntLiteral(bytes []byte) (v int64, err error) {
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
		case c == UnderscoreToken && base0:
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
	if len(s) >= 1 && (s[0] == MinusToken || s[0] == PlusToken) {
		s = s[1:]
	}

	// Optional base prefix.
	hex := false
	if len(s) >= 2 && s[0] == '0' && (lower(s[1]) == 'b' || lower(s[1]) == 'o' || lower(s[1]) == 'x') {
		i = 2
		saw = '0' // base prefix counts as a digit for "underscore as digit separator"
		hex = lower(s[1]) == 'x'
	}

	for ; i < len(s); i++ {
		currentChar := s[i]
		isDigit := '0' <= currentChar && currentChar <= '9'
		isHexChar := hex && 'a' <= lower(currentChar) && lower(currentChar) <= 'f'

		if isDigit || isHexChar {
			saw = '0'
		} else if currentChar == UnderscoreToken {
			if saw != '0' {
				return false
			}
			saw = UnderscoreToken
		} else {
			if saw == UnderscoreToken {
				return false
			}
			saw = BangToken
		}
	}

	return saw != UnderscoreToken
}

// getType is a function that takes a byte slice and an offset as input parameters.
// It returns a byte slice representing the data type, a ValueType indicating the type of data,
// an integer representing the end offset, and an error if any.
// If the input byte slice is empty, it returns an error indicating that no JSON data was provided.
// Otherwise, it calls the parseValue function to parse the value and returns the parsed data type,
// the end offset, and any error encountered during parsing.
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

// parseValue parses a JSON value from the given data starting at the specified offset.
// It returns the parsed value, the new offset after parsing, and an error if any.
func parseValue(data []byte, offset int) (ValueType, int, error) {
	switch data[offset] {
	case DoublyQuoteToken:
		return parseString(data, offset)
	case ArrayStartToken, ObjectStartToken:
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

	if data[offset] == ArrayStartToken {
		containerType = Array
		closing = ArrayEndToken
	} else {
		containerType = Object
		closing = ObjectEndToken
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
		if _boolValue, err := ParseBoolLiteral(value); err != nil {
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
			if _intValue, err := ParseIntLiteral(value); err != nil {
				return Number, offset + end, err
			}
		}

		if _floatValue, err := ParseFloatLiteral(value); err != nil {
			return Number, offset + end, err
		}

		return Number, offset + end, nil
	}

	return Unknown, offset, errors.New("unknown value type")
}

func ArrayEach(data []byte, callback func(value []byte, dataType ValueType, offset int, err error), keys ...string) (int, error) {
	if len(data) == 0 {
		return -1, MalformedObject
	}

	offset, err := parseInitialOffset(data, keys)
	if err != nil {
		return -1, err
	}

	return parseArrayElements(data, offset, callback)
}

func parseInitialOffset(data []byte, keys []string) (int, error) {
	offset, err := nextToken(data)
	if err != nil {
		return -1, MalformedJson
	}

	if len(keys) > 0 {
		offset, err = findValueIndex(data, keys...)
		if err != nil {
			return -1, KeyPathNotFoundError
		}

		if offset, err = parseUntilArrayStart(data, offset); err != nil {
			return -1, err
		}
	}

	return offset, nil
}

func parseUntilArrayStart(data []byte, offset int) (int, error) {
	offset, err := nextToken(data[offset:])
	if err != nil {
		return -1, MalformedJson
	}

	if data[offset] != ArrayStartToken {
		return -1, MalformedArray
	}

	return offset + 1, nil
}

func parseArrayElements(data []byte, offset int, f func(value []byte, dataType ValueType, offset int, err error)) (int, error) {
	for data[offset] != ArrayEndToken {
		value, valueType, valueOffset, err := Get(data[offset:])
		if err != nil {
			return offset, err
		}

		if valueOffset == 0 || valueType == NotExist {
			break
		}

		f(value, valueType, offset+valueOffset-len(value), err)

		if offset, err = parseNextElementOrEndArray(data, offset+valueOffset); err != nil {
			return offset, err
		}
	}

	return offset, nil
}

func parseNextElementOrEndArray(data []byte, offset int) (int, error) {
	offset, err := nextToken(data[offset:])
	if err != nil {
		return -1, MalformedArray
	}

	if data[offset] != ValueSeparatorToken && data[offset] != ArrayEndToken {
		return -1, MalformedArray
	}

	return offset + 1, nil
}
