// Regression test: bigdec → float32 must round once at 24-bit precision,
// not twice via float64. The value below exceeds the float32 halfway point
// between 1.0 (0x3f800000) and the next representable float32 (0x3f800001)
// only because of the 0x1p-80 term, which vanishes when the value is first
// rounded to float64 (whose ULP at 1.0 is 2^-52). Direct rounding sees the
// value as strictly above the halfway point and rounds up to 0x3f800001;
// double rounding sees it as exactly halfway and rounds to even = 0x3f800000.

package main

import "math"

func main() {
	const x float32 = 1.0 + 0x1p-24 + 0x1p-80
	println("value of x:", x)
	println("bits of x:", math.Float32bits(x))
	println("next float32:", math.Float32bits(math.Nextafter32(x, 1e10)))
}

// Output:
// value of x: 1.0000001
// bits of x: 1065353217
// next float32: 1065353218
