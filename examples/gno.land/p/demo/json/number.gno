package json

import (
	"errors"
	"math"
	"strconv"
)

const (
	EmptyBytes error	 = errors.New("empty bytes")
	InvalidExponents	 = errors.New("invalid exponents")
	NonDigitCharacters	 = errors.New("non-digit characters")
	Overflow			 = errors.New("overflow")
	MultipleDecimalPoints = errors.New("multiple decimal points")
)

const (
	absMinInt64 = 1 << 63
	maxInt64    = 1<<63 - 1
	maxUint64   = 1<<64 - 1
)

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
		if c < '0' || c > '9' {
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

		if c < '0' || c > '9' {
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
