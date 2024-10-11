package json

import (
	"bytes"
	"errors"
	"strconv"

	el "gno.land/p/demo/json/eisel_lemire"
)

const (
	absMinInt64 = 1 << 63
	maxInt64    = absMinInt64 - 1
	maxUint64   = 1<<64 - 1
)

const unescapeStackBufSize = 64

const (
	errInvalidStringInput    = errors.New("invalid string input")
	errMalformedBooleanValue = errors.New("malformed boolean value")
	errEmptyByteSlice        = errors.New("empty byte slice")
	errInvalidExponentValue  = errors.New("invalid exponent value")
	errNonDigitCharacters    = errors.New("non-digit characters found")
	errNumericRangeExceeded  = errors.New("numeric value exceeds the range limit")
	errMultipleDecimalPoints = errors.New("multiple decimal points found")
)

// PaseStringLiteral parses a string from the given byte slice.
func ParseStringLiteral(data []byte) (string, error) {
	var buf [unescapeStackBufSize]byte

	bf, err := Unescape(data, buf[:])
	if err != nil {
		return "", errInvalidStringInput
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
		return false, errMalformedBooleanValue
	}
}

// PaseFloatLiteral parses a float64 from the given byte slice.
//
// It utilizes double-precision (64-bit) floating-point format as defined
// by the IEEE 754 standard, providing a decimal precision of approximately 15 digits.
func ParseFloatLiteral(bytes []byte) (float64, error) {
	if len(bytes) == 0 {
		return -1, errEmptyByteSlice
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
			return -1, errInvalidExponentValue
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

func ParseIntLiteral(bytes []byte) (int64, error) {
	if len(bytes) == 0 {
		return 0, errEmptyByteSlice
	}

	neg, bytes := trimNegativeSign(bytes)

	var n uint64 = 0
	for _, c := range bytes {
		if notDigit(c) {
			return 0, errNonDigitCharacters
		}

		if n > maxUint64/10 {
			return 0, errNumericRangeExceeded
		}

		n *= 10

		n1 := n + uint64(c-'0')
		if n1 < n {
			return 0, errNumericRangeExceeded
		}

		n = n1
	}

	if n > maxInt64 {
		if neg && n == absMinInt64 {
			return -absMinInt64, nil
		}

		return 0, errNumericRangeExceeded
	}

	if neg {
		return -int64(n), nil
	}

	return int64(n), nil
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
				return 0, 0, errMultipleDecimalPoints
			}
			decimalFound = true
			continue
		}

		if notDigit(c) {
			return 0, 0, errNonDigitCharacters
		}

		digit := uint64(c - '0')

		if man > (maxUint64-digit)/10 {
			return 0, 0, errNumericRangeExceeded
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

func notDigit(c byte) bool {
	return (c & 0xF0) != 0x30
}

// lower converts a byte to lower case if it is an uppercase letter.
func lower(c byte) byte {
	return c | 0x20
}
