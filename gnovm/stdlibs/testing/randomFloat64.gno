package testing

import (
	"math"
	"time"
)

func randomFloat32(a float32) float32 {
	bits := math.Float32bits(a)

	exponent := (bits >> 23) & 0xFF
	mantissa := bits & 0x7FFFFF
	sign := bits & 0x80000000
	t := uint32(unixNano())
	manshift := 1 + (t % 7)

	var shift int8
	if exponent <= 1 {
		shift = int8(1 + int(mantissa%2))
	} else if exponent >= 0xFE {
		shift = int8(-1 - int(mantissa%2))
	} else {
		shift = int8(-2 + int(mantissa%5))
	}

	newExp := int32(exponent) + int32(shift)
	newExponent := uint32(newExp)

	newMantissa := mantissa ^ (mantissa >> manshift)

	newBits := sign | (newExponent << 23) | (newMantissa & 0x7FFFFF)

	return math.Float32frombits(newBits)
}

func randomFloat64(a float64) float64 {
	bits := math.Float64bits(a)

	exponent := (bits >> 52) & 0x7FF

	mantissa := bits & 0xFFFFFFFFFFFFF

	sign := bits & 0x8000000000000000

	t := uint64(time.Now().UnixNano())
	manshift := 1 + (t % 7)

	var shift int16
	if exponent <= 1 {
		shift = int16(1 + int64(mantissa%2))
	} else if exponent >= 0x7FE {
		shift = int16(-1 - int64(mantissa%2))
	} else {
		shift = int16(-2 + int64(mantissa%5))
	}

	newExp := int64(exponent) + int64(shift)
	newExponent := uint64(newExp)

	newMantissa := mantissa ^ (mantissa >> manshift)

	newBits := sign | (newExponent << 52) | (newMantissa & 0xFFFFFFFFFFFFF)

	return math.Float64frombits(newBits)
}
