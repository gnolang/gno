// PKGPATH: gno.land/r/ftoatest
package ftoatest

// PROBE: does a /r/ realm calling strconv.FormatFloat on extreme/denormal
// floats (deep ryu path -> mult128bitPow10 q<0 -> `pow[0] += 1`) hit the
// readonly taint? If it panics, the stdlib self-table copy-mutate bug is
// latent at HEAD independent of the /p/+stdlib realm work.

import (
	"math"
	"strconv"
)

func main(cur realm) {
	vals := []float64{
		5e-324,                     // smallest denormal
		math.SmallestNonzeroFloat64,
		2.2250738585072014e-308,    // smallest normal
		1.7976931348623157e+308,    // max
		0.30000000000000004,        // 0.1+0.2
		1.0 / 3.0,
		9.999999999999999e22,
		1e-300, 7.3e-200, 4.9e-100,
		math.Float64frombits(0x0123456789abcdef),
		math.Float64frombits(0xfedcba9876543210),
	}
	for _, v := range vals {
		_ = strconv.FormatFloat(v, 'g', -1, 64)
	}
	println("ok, no panic")
}

// Output:
// ok, no panic
