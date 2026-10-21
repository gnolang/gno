// run

// Transcribed from Go's go/constant tests and issues #6389 #6392 #6393 #6394; expected output from go1.25.9.
// Integer constants convert to float32 and float64 exactly when they fit
// the mantissa, with large exponents handled in exact arithmetic.

package main

const maxFloat32 = 1 << 127 * (1<<24 - 1) / (1.0 << 23)
const smallestFloat32 = 1.0 / (1 << (127 - 1 + 23))
const smallestFloat64 = 1.0 / (1 << 500) / (1 << 500) / (1 << 74)

func main() {
	println(int64(float64(1<<53)), int64(float64(1<<62)))
	println(int64(float32(1<<24)), int64(float32(1<<62)))
	println(float64(1<<100) == 1<<100, float32(1<<100) == 1<<100)
	println(float64(1<<63) == 9223372036854775808, float64(1<<64) == 18446744073709551616)
	println(float64(-1<<63) == -9223372036854775808)
	println(float64(1<<200)/float64(1<<100) == 1<<100)

	println(float32(maxFloat32) == maxFloat32, float32(-maxFloat32) == -maxFloat32)
	println(float32(smallestFloat32) == smallestFloat32, float32(smallestFloat32) > 0)
	println(float32(smallestFloat32/2) == 0, float32(smallestFloat64) == 0)
	println(float64(smallestFloat32) == smallestFloat32, float64(smallestFloat32/2) == smallestFloat32/2)
	println(float64(smallestFloat64) > 0, float64(smallestFloat64/2) == 0)

	var h float64 = 1 << 62
	var i float32 = 1 << 24
	var j float64 = 123456789012345678
	println(int64(h), int64(i), int64(j))
	println(float64(1e300)/1e299 == 10, 1e100+1e100 == 2e100, 1e100-1e100 == 0)
}

// GnoOutput:
// 9007199254740992 4611686018427387904
// 16777216 4611686018427387904
// true true
// true true
// true
// true
// true true
// true true
// true true
// true true
// true true
// 4611686018427387904 16777216 123456789012345680
// true true true

// GoOutput:
// 9007199254740992 4611686018427387904
// 16777216 4611686018427387904
// true true
// true true
// true
// true
// true true
// true true
// true true
// true true
// true true
// 4611686018427387904 16777216 123456789012345680
// true true true

// Verdict: Clean run
