// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that incorrect invocations of the complex predeclared function are detected.
// Does not compile.

package main

type (
	Float32    float32
	Float64    float64
	Complex64  complex64
	Complex128 complex128
)

var (
	f32 float32
	f64 float64
	F32 Float32
	F64 Float64

	c64  complex64
	c128 complex128
	C64  Complex64
	C128 Complex128
)

func F1() int {
	return 1
}

func F3() (int, int, int) {
	return 1, 2, 3
}

func main() {
	// ok
	c64 = complex(f32, f32)
	c128 = complex(f64, f64)

	_ = complex128(0)     // ok
	_ = complex(f32, f64) // ERROR "complex"
	_ = complex(f64, f32) // ERROR "complex"
	_ = complex(f32, F32) // ERROR "complex"
	_ = complex(F32, f32) // ERROR "complex"
	_ = complex(f64, F64) // ERROR "complex"
	_ = complex(F64, f64) // ERROR "complex"

	_ = complex(F1()) // ERROR "not enough arguments"
	_ = complex(F3()) // ERROR "too many arguments"

	_ = complex() // ERROR "not enough arguments"

	c128 = complex(f32, f32) // ERROR "cannot use"
	c64 = complex(f64, f64)  // ERROR "cannot use"

	c64 = complex(1.0, 2.0) // ok, constant is untyped
	c128 = complex(1.0, 2.0)
	C64 = complex(1.0, 2.0)
	C128 = complex(1.0, 2.0)

	C64 = complex(f32, f32)  // ERROR "cannot use"
	C128 = complex(f64, f64) // ERROR "cannot use"

}

// GoTypeCheckError:
// line 45: invalid operation: complex(f32, f64) (mismatched types float32 and float64)
// line 46: invalid operation: complex(f64, f32) (mismatched types float64 and float32)
// line 47: invalid operation: complex(f32, F32) (mismatched types float32 and Float32)
// line 48: invalid operation: complex(F32, f32) (mismatched types Float32 and float32)
// line 49: invalid operation: complex(f64, F64) (mismatched types float64 and Float64)
// line 50: invalid operation: complex(F64, f64) (mismatched types Float64 and float64)
// line 52: invalid operation: not enough arguments for complex(F1()) (expected 2, found 1)
// line 53: invalid operation: too many arguments for complex(F3()) (expected 2, found 3)
// line 55: invalid operation: not enough arguments for complex() (expected 2, found 0)
// line 57: cannot use complex(f32, f32) (value of type complex64) as complex128 value in assignment
// line 58: cannot use complex(f64, f64) (value of type complex128) as complex64 value in assignment
// line 65: cannot use complex(f32, f32) (value of type complex64) as Complex64 value in assignment
// line 66: cannot use complex(f64, f64) (value of type complex128) as Complex128 value in assignment

// KnownIssue:
// line 15: name complex64 not defined in fileset with files [cmplx.go]
