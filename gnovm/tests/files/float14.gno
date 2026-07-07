// Constant expressions have no signed zero: negating a zero constant
// (untyped or typed) or multiplying one by -1 yields +0; only runtime
// negation of a zero variable yields -0. Mirrors Go corpus
// fixedbugs/bug434.go.

package main

import "math"

const zero = 0.0

const zt float64 = 0.0

func bits(f float64) string {
	switch math.Float64bits(f) {
	case 0:
		return "+0"
	case 1 << 63:
		return "-0"
	}
	return "nonzero"
}

func main() {
	x := -zero
	println("untyped -zero:", bits(x))
	x = -float64(zero)
	println("typed -conv:", bits(x))
	v := x
	println("runtime -v:", bits(-v))
	println("typed const arith:", bits(zt*-1), bits(-1*zt))
}

// Output:
// untyped -zero: +0
// typed -conv: +0
// runtime -v: -0
// typed const arith: +0 +0
