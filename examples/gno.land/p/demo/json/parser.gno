package json

import (
	"bytes"
	"errors"
	"math"
	"strconv"
)

func tokenEnd(data []byte) int {
	for i, tok := range data {
		switch tok {
		case WhiteSpaceToken, NewLineToken, CarriageReturnToken, TabToken, CommaToken, CurlyCloseToken, SquareCloseToken:
			return i
		}
	}

	return len(data)
}

func nextToken(data []byte) (int, error) {
	for i, c := range data {
		switch c {
		case WhiteSpaceToken, NewLineToken, CarriageReturnToken, TabToken:
			continue
		default:
			return i, nil
		}
	}

	return -1, TokenNotFound
}

func stringEnd(data []byte) (lastCharIdx int, escaped bool) {
	escaped = false
	for lastCharIdx, c := range data {
		switch c {
		case DoublyQuoteToken:
			if !escaped {
				return lastCharIdx + 1, false
			}

			prevCharIdx := lastCharIdx - 1
			for prevCharIdx >= 0 && data[prevCharIdx] == BackSlashToken {
				prevCharIdx--
			}

			if (lastCharIdx-prevCharIdx)%2 == 0 {
				return lastCharIdx + 1, true
			}

			return lastCharIdx + 1, false
		case BackSlashToken:
			escaped = true
		}
	}

	return -1, escaped
}

// Find end of the data structure, array or object.
// For array openSym and closeSym will be '[' and ']', for object '{' and '}'
func blockEnd(data []byte, openSym byte, closeSym byte) int {
	level := 0

	for i := 0; i < len(data); i++ {
		switch data[i] {
		case DoublyQuoteToken:
			se, _ := stringEnd(data[i+1:])
			if se == -1 {
				return -1
			}
			i += se
		case openSym:
			level += 1
		case closeSym:
			level -= 1
			if level == 0 {
				return i + 1
			}
		}
	}

	return -1
}

func keyMatched(
	key []byte,
	keyEscaped bool,
	stackbuf [unescapeStackBufSize]byte,
	keys []string, level int,
) (keyUnesc []byte, err error) {
	if !keyEscaped {
		keyUnesc = key
	}

	if ku, err := unescape(key, stackbuf[:]); err != nil {
		return nil, err
	} else {
		keyUnesc = ku
	}

	if level > len(keys) {
		return nil, KeyLevelNotMatched
	}

	return keyUnesc, err
}

// PaseStringLiteral parses a string from the given byte slice.
func ParseStringLiteral(data []byte) (string, error) {
	var buf [unescapeStackBufSize]byte

	bf, err := unescape(data, buf[:])
	if err != nil {
		return "", MalformedString
	}

	return string(bf), nil
}

// ParseBoolLiteral parses a boolean value from the given byte slice.
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

// PaseFloatLiteral parses a float64 from the given byte slice.
//
// It utilizes double-precision (64-bit) floating-point format as defined
// by the IEEE 754 standard, providing a decimal precision of approximately 15 digits.
// TODO: support for 32-bit floating-point format
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

	// for fast float64 conversion
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
		} else if currentChar == UnderScoreToken {
			if saw != '0' {
				return false
			}
			saw = UnderScoreToken
		} else {
			if saw == UnderScoreToken {
				return false
			}
			saw = BangToken
		}
	}

	return saw != UnderScoreToken
}

// getTypeFromByteSlice is a function that takes a byte slice and an offset as input parameters.
// It returns a byte slice representing the data type, a ValueType indicating the type of data,
// an integer representing the end offset, and an error if any.
// If the input byte slice is empty, it returns an error indicating that no JSON data was provided.
// Otherwise, it calls the parseValue function to parse the value and returns the parsed data type,
// the end offset, and any error encountered during parsing.
func getTypeFromByteSlice(data []byte, offset int) ([]byte, ValueType, int, error) {
	if len(data) == 0 {
		return nil, Unknown, offset, errors.New("no JSON data provided")
	}

	dataType, endOffset, err := parseValue(data, offset)
	if err != nil {
		return nil, dataType, offset, err
	}

	return data[offset:endOffset], dataType, endOffset, nil
}

// getTypeFromValue check the type of the given value and returns the corresponding ValueType.
func getTypeFromValue(v interface{}) ValueType {
	switch v.(type) {
	case string:
		return String
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return Number
	case float32, float64:
		return Float
	case bool:
		return Boolean
	case nil:
		return Null
	case map[string]interface{}:
		return Object
	case []interface{}:
		return Array
	default:
		return Unknown
	}
}

// parseValue parses a JSON value from the given data starting at the specified offset.
// It returns the parsed value, the new offset after parsing, and an error if any.
func parseValue(data []byte, offset int) (ValueType, int, error) {
	switch data[offset] {
	case DoublyQuoteToken:
		return parseString(data, offset)
	case SquareOpenToken, CurlyOpenToken:
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

	if data[offset] == SquareOpenToken {
		containerType = Array
		closing = SquareCloseToken
	} else {
		containerType = Object
		closing = CurlyCloseToken
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
