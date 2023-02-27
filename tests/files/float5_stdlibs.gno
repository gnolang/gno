package main

import (
	"math"

	imath "internal/math"
)

func main() {
	println(math.MaxFloat32, imath.Float32bits(math.MaxFloat32))
	println(math.MaxFloat64, imath.Float64bits(math.MaxFloat64))
}

// Output:
// 3.4028234663852886e+38 2139095039
// 1.7976931348623157e+308 9218868437227405311

// NOTE: 0x7f7fffff is 2139095039
// NOTE: 0x7fefffffffffffff is 9218868437227405311
