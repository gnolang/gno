// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import (
	imath "internal/math"
)

// Nextafter32 returns the next representable float32 value after x towards y.
//
// Special cases are:
//
//	Nextafter32(x, x)   = x
//	Nextafter32(NaN, y) = NaN
//	Nextafter32(x, NaN) = NaN
func Nextafter32(x, y float32) (r float32) {
	switch {
	case IsNaN(float64(x)) || IsNaN(float64(y)): // special case
		r = float32(NaN())
	case x == y:
		r = x
	case x == 0:
		r = float32(Copysign(float64(imath.Float32frombits(1)), float64(y)))
	case (y > x) == (x > 0):
		r = imath.Float32frombits(imath.Float32bits(x) + 1)
	default:
		r = imath.Float32frombits(imath.Float32bits(x) - 1)
	}
	return
}

// Nextafter returns the next representable float64 value after x towards y.
//
// Special cases are:
//
//	Nextafter(x, x)   = x
//	Nextafter(NaN, y) = NaN
//	Nextafter(x, NaN) = NaN
func Nextafter(x, y float64) (r float64) {
	switch {
	case IsNaN(x) || IsNaN(y): // special case
		r = NaN()
	case x == y:
		r = x
	case x == 0:
		r = Copysign(imath.Float64frombits(1), y)
	case (y > x) == (x > 0):
		r = imath.Float64frombits(imath.Float64bits(x) + 1)
	default:
		r = imath.Float64frombits(imath.Float64bits(x) - 1)
	}
	return
}
