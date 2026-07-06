// A negative float constant that underflows must convert to +0, not -0:
// constants have no signed zero (golang/go#12576, golang/go#12621).
// Runtime conversions and arithmetic keep the sign of an underflowed zero.

package main

import "math"

var m = -1e-10000

const tiny float64 = -1e-50

func main() {
	println("var float64:", math.Signbit(m))
	var f32 float32 = -1e-50
	println("var float32:", math.Signbit(float64(f32)))
	println("conv float32:", math.Signbit(float64(float32(-1e-50))))
	println("conv typed const:", math.Signbit(float64(float32(tiny))))
	var i interface{} = -1e-10000
	println("interface:", math.Signbit(i.(float64)))
	d := -math.SmallestNonzeroFloat64
	println("runtime narrowing:", math.Signbit(float64(float32(d))))
	println("runtime underflow:", math.Signbit(d/3))
}

// Output:
// var float64: false
// var float32: false
// conv float32: false
// conv typed const: false
// interface: false
// runtime narrowing: true
// runtime underflow: true
