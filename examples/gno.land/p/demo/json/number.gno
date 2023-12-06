package json

import (
	"encoding/binary"
	"math"
	"strconv"
)

const (
	absMinInt64 = 1 << 63
	maxInt64    = 1<<63 - 1
	maxUint64   = 1<<64 - 1
)

func ParseNumberKind(bytes []byte) (value float64, isFloat, ok, overflow bool) {
    if len(bytes) == 0 {
        return 0, false, false, false
    }

    neg := bytes[0] == '-'
    if neg {
        bytes = bytes[1:]
    }

    var intPart, fracPart uint64
    var decimalFound bool
    var fracDivisor uint64 = 1

    for _, c := range bytes {
        if c == '.' {
            if decimalFound {
                return 0, false, false, false
            }
            decimalFound = true
            continue
        }

        if c < '0' || c > '9' {
            return 0, false, false, false
        }

        digit := uint64(c - '0')
        if !decimalFound {
            if intPart > (maxUint64-uint64(digit))/10 {
                return 0, false, false, true
            }
            intPart = intPart*10 + digit
        } else {
            if fracPart > (maxUint64-uint64(digit))/10 {
                return 0, false, false, true
            }
            fracPart = fracPart*10 + digit
            fracDivisor *= 10
        }
    }

    if decimalFound {
        isFloat = true
    }

    totalValue := float64(intPart) + float64(fracPart)/float64(fracDivisor)
    if neg {
        totalValue = -totalValue
    }

    // Overflow check for integer part
    if !isFloat {
        if neg {
            if intPart > absMinInt64 {
                return 0, false, false, true
            }
        } else {
            if intPart > maxInt64 {
                return 0, false, false, true
            }
        }

		// return totalValue instead of float64(intPart)
        return totalValue, isFloat, true, false
    }

    return totalValue, isFloat, true, false
}