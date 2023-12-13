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
)

func ParseString(data []byte) (string, error) {
	var buf [UnescapeStackBufSize]byte

	bf, err := Unescape(data, buf[:])
	if err != nil {
		return "", MalformedStringError
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
		return false, MalformedValueError
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
		return Unknown, offset, UnknownValueTypeError
	case 'u', 'n':
		if bytes.Equal(value, NullLiteral) {
			dataType = Null
		}
		return Unknown, offset, UnknownValueTypeError
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		dataType = Number
	default:
		return Unknown, offset, UnknownValueTypeError
	}
}