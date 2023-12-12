package json

import (
	"errors"
	"math"
	"strconv"
)

const (
	absMinInt64 = 1 << 63
	maxInt64    = 1<<63 - 1
	maxUint64   = 1<<64 - 1
)

func ParseFloat(bytes []byte) (value float64, isFloat, ok bool) {
	if len(bytes) == 0 {
		panic("invalid number: empty string")
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

	man, exp10 := extractMantissaAndExp10(bytes)

	if len(exponentPart) > 0 {
		exp, err := strconv.Atoi(string(exponentPart))
		if err != nil {
			panic("invalid exponent")
		}
		exp10 += exp
	}

	f, success := eiselLemire64(man, exp10, neg)
	if !success {
		return 0, false, false
	}

	return f, true, true
}

func ParseInt(bytes []byte) (v int64, ok, overflow bool) {
	if len(bytes) == 0 {
		return 0, false, false
	}

	neg, bytes := trimNegativeSign(bytes)

	var n uint64 = 0
	for _, c := range bytes {
		if c < '0' || c > '9' {
			return 0, false, false
		}

		if n > maxUint64/10 {
			return 0, false, true
		}

		n *= 10

		n1 := n + uint64(c-'0')
		if n1 < n {
			return 0, false, true
		}

		n = n1
	}

	if n > maxInt64 {
		if neg && n == absMinInt64 {
			return -absMinInt64, true, false
		}

		return 0, false, true
	}

	if neg {
		return -int64(n), true, false
	}

	return int64(n), true, false
}

func trimNegativeSign(bytes []byte) (neg bool, trimmed []byte) {
	if bytes[0] == '-' {
		return true, bytes[1:]
	}

	return false, bytes
}

func extractMantissaAndExp10(bytes []byte) (uint64, int) {
	var man uint64
	var exp10 int
	decimalFound := false

	for _, c := range bytes {
		if c == '.' {
			if decimalFound {
				panic("invalid number: multiple decimal points")
			}
			decimalFound = true
			continue
		}

		if c < '0' || c > '9' {
			panic("invalid number: non-digit characters")
		}

		digit := uint64(c - '0')

		if man > (maxUint64-digit)/10 {
			panic("invalid number: overflow")
		}

		man = man*10 + digit

		if decimalFound {
			exp10--
		}
	}

	return man, exp10
}
