package json

import (
	"bytes"
	"errors"
	"strconv"
	"unicode"

	el "gno.land/p/demo/json/eisel_lemire"
)

const (
	absMinInt64 = 1 << 63
	maxInt64    = absMinInt64 - 1
	maxUint64   = 1<<64 - 1
)

const unescapeStackBufSize = 64

// PaseStringLiteral parses a string from the given byte slice.
func ParseStringLiteral(data []byte) (string, error) {
	var buf [unescapeStackBufSize]byte

	bf, err := Unescape(data, buf[:])
	if err != nil {
		return "", errors.New("invalid string input found while parsing string value")
	}

	return string(bf), nil
}

// ParseBoolLiteral parses a boolean value from the given byte slice.
func ParseBoolLiteral(data []byte) (bool, error) {
	switch {
	case bytes.Equal(data, trueLiteral):
		return true, nil
	case bytes.Equal(data, falseLiteral):
		return false, nil
	default:
		return false, errors.New("JSON Error: malformed boolean value found while parsing boolean value")
	}
}

// PaseFloatLiteral parses a float64 from the given byte slice.
//
// It utilizes double-precision (64-bit) floating-point format as defined
// by the IEEE 754 standard, providing a decimal precision of approximately 15 digits.
func ParseFloatLiteral(bytes []byte) (float64, error) {
	if len(bytes) == 0 {
		return -1, errors.New("JSON Error: empty byte slice found while parsing float value")
	}

	neg, bytes := trimNegativeSign(bytes)

	var exponentPart []byte
	for i, c := range bytes {
		if lower(c) == 'e' {
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
			return -1, errors.New("JSON Error: invalid exponent value found while parsing float value")
		}
		exp10 += exp
	}

	// for fast float64 conversion
	f, success := el.EiselLemire64(man, exp10, neg)
	if !success {
		return 0, nil
	}

	return f, nil
}

// extractMantissaAndExp10 parses a byte slice representing a decimal number and extracts the mantissa and the exponent of its base-10 representation.
// It iterates through the bytes, constructing the mantissa by treating each byte as a digit.
// If a decimal point is encountered, the function keeps track of the position of the decimal point to calculate the exponent.
// The function ensures that:
// - The number contains at most one decimal point.
// - All characters in the byte slice are digits or a single decimal point.
// - The resulting mantissa does not overflow a uint64.
func extractMantissaAndExp10(bytes []byte) (uint64, int, error) {
	var (
		man          uint64
		exp10        int
		decimalFound bool
	)

	for _, c := range bytes {
		if c == dot {
			if decimalFound {
				return 0, 0, errors.New("JSON Error: multiple decimal points found while parsing float value")
			}
			decimalFound = true
			continue
		}

		if !unicode.IsDigit(rune(c)) {
			return 0, 0, errors.New("JSON Error: non-digit characters found while parsing integer value")
		}

		digit := uint64(c - '0')

		if man > (maxUint64-digit)/10 {
			return 0, 0, errors.New("JSON Error: numeric value exceeds the range limit")
		}

		man = man*10 + digit

		if decimalFound {
			exp10--
		}
	}

	return man, exp10, nil
}

func trimNegativeSign(bytes []byte) (bool, []byte) {
	if bytes[0] == minus {
		return true, bytes[1:]
	}

	return false, bytes
}

// lower converts a byte to lower case if it is an uppercase letter.
func lower(c byte) byte {
	return c | 0x20
}

func isAlphaNumeric(c byte) bool {
	return unicode.IsLetter(rune(c)) || unicode.IsDigit(rune(c))
}
