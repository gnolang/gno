// A float constant that underflows must convert to +0, not -0: constants
// have no signed zero (golang/go#12576, golang/go#12621). Runtime
// conversions and arithmetic keep the sign of an underflowed zero.

package main

import "math"

var m = -1e-10000

const tiny float64 = -1e-50

func zero(f float64) string {
	if math.Signbit(f) {
		return "-0"
	}
	return "+0"
}

func main() {
	println("var float64:", zero(m))
	var f32 float32 = -1e-50
	println("var float32:", zero(float64(f32)))
	println("conv float32:", zero(float64(float32(-1e-50))))
	println("conv typed const:", zero(float64(float32(tiny))))
	var i interface{} = -1e-10000
	println("interface:", zero(i.(float64)))
	d := -math.SmallestNonzeroFloat64
	println("runtime narrowing:", zero(float64(float32(d))))
	println("runtime underflow:", zero(d/3))
	// Nonzero constants are unaffected by the constant-space narrowing.
	var v float32 = 0.3
	println("nonzero conv:", float32(0.3) == v, float32(tiny) == 0)
}

// Output:
// var float64: +0
// var float32: +0
// conv float32: +0
// conv typed const: +0
// interface: +0
// runtime narrowing: -0
// runtime underflow: -0
// nonzero conv: true true
