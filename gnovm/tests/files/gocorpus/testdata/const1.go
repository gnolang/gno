// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify overflow is detected when using numeric constants.
// Does not compile.

package main

import "unsafe"

type I interface{}

const (
	// assume all types behave similarly to int8/uint8
	Int8   int8  = 101
	Minus1 int8  = -1
	Uint8  uint8 = 102
	Const        = 103

	Float32    float32 = 104.5
	Float64    float64 = 105.5
	ConstFloat         = 106.5
	Big        float64 = 1e300

	String = "abc"
	Bool   = true
)

var (
	a1 = Int8 * 100              // ERROR "overflow|cannot convert"
	a2 = Int8 * -1               // OK
	a3 = Int8 * 1000             // ERROR "overflow|cannot convert"
	a4 = Int8 * int8(1000)       // ERROR "overflow|cannot convert"
	a5 = int8(Int8 * 1000)       // ERROR "overflow|cannot convert"
	a6 = int8(Int8 * int8(1000)) // ERROR "overflow|cannot convert"
	a7 = Int8 - 2*Int8 - 2*Int8  // ERROR "overflow|cannot convert"
	a8 = Int8 * Const / 100      // ERROR "overflow|cannot convert"
	a9 = Int8 * (Const / 100)    // OK

	b1        = Uint8 * Uint8         // ERROR "overflow|cannot convert"
	b2        = Uint8 * -1            // ERROR "overflow|cannot convert"
	b3        = Uint8 - Uint8         // OK
	b4        = Uint8 - Uint8 - Uint8 // ERROR "overflow|cannot convert"
	b5        = uint8(^0)             // ERROR "overflow|cannot convert"
	b5a       = int64(^0)             // OK
	b6        = ^uint8(0)             // OK
	b6a       = ^int64(0)             // OK
	b7        = uint8(Minus1)         // ERROR "overflow|cannot convert"
	b8        = uint8(int8(-1))       // ERROR "overflow|cannot convert"
	b8a       = uint8(-1)             // ERROR "overflow|cannot convert"
	b9   byte = (1 << 10) >> 8        // OK
	b10  byte = (1 << 10)             // ERROR "overflow|cannot convert"
	b11  byte = (byte(1) << 10) >> 8  // ERROR "overflow|cannot convert"
	b12  byte = 1000                  // ERROR "overflow|cannot convert"
	b13  byte = byte(1000)            // ERROR "overflow|cannot convert"
	b14  byte = byte(100) * byte(100) // ERROR "overflow|cannot convert"
	b15  byte = byte(100) * 100       // ERROR "overflow|cannot convert"
	b16  byte = byte(0) * 1000        // ERROR "overflow|cannot convert"
	b16a byte = 0 * 1000              // OK
	b17  byte = byte(0) * byte(1000)  // ERROR "overflow|cannot convert"
	b18  byte = Uint8 / 0             // ERROR "division by zero"

	c1 float64 = Big
	c2 float64 = Big * Big          // ERROR "overflow|cannot convert"
	c3 float64 = float64(Big) * Big // ERROR "overflow|cannot convert"
	c4         = Big * Big          // ERROR "overflow|cannot convert"
	c5         = Big / 0            // ERROR "division by zero"
	c6         = 1000 % 1e3         // ERROR "invalid operation|expected integer type"
)

func f(int)

func main() {
	f(Int8)             // ERROR "convert|wrong type|cannot"
	f(Minus1)           // ERROR "convert|wrong type|cannot"
	f(Uint8)            // ERROR "convert|wrong type|cannot"
	f(Const)            // OK
	f(Float32)          // ERROR "convert|wrong type|cannot"
	f(Float64)          // ERROR "convert|wrong type|cannot"
	f(ConstFloat)       // ERROR "truncate"
	f(ConstFloat - 0.5) // OK
	f(Big)              // ERROR "convert|wrong type|cannot"
	f(String)           // ERROR "convert|wrong type|cannot|incompatible"
	f(Bool)             // ERROR "convert|wrong type|cannot|incompatible"
}

const ptr = nil // ERROR "const.*nil|not constant"
const _ = string([]byte(nil)) // ERROR "is not a? ?constant"
const _ = uintptr(unsafe.Pointer((*int)(nil))) // ERROR "is not a? ?constant"
const _ = unsafe.Pointer((*int)(nil)) // ERROR "cannot be nil|invalid constant type|is not a constant|not constant"
const _ = (*int)(nil) // ERROR "cannot be nil|invalid constant type|is not a constant|not constant"

// GnoIncomplete: covered 40 of 42 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// GnoError:
// line 12: unknown import path unsafe

// GoTypeCheckError:
// line 33: Int8 * 100 (constant 10100 of type int8) overflows int8
// line 35: 1000 (untyped int constant) overflows int8
// line 36: constant 1000 overflows int8
// line 37: 1000 (untyped int constant) overflows int8
// line 38: constant 1000 overflows int8
// line 39: 2 * Int8 (constant 202 of type int8) overflows int8
// line 40: Int8 * Const (constant 10403 of type int8) overflows int8
// line 43: Uint8 * Uint8 (constant 10404 of type uint8) overflows uint8
// line 44: -1 (untyped int constant) overflows uint8
// line 46: Uint8 - Uint8 - Uint8 (constant -102 of type uint8) overflows uint8
// line 47: constant -1 overflows uint8
// line 51: constant -1 overflows uint8
// line 52: constant -1 overflows uint8
// line 53: constant -1 overflows uint8
// line 55: cannot use (1 << 10) (untyped int constant 1024) as byte value in variable declaration (overflows)
// line 56: byte(1) << 10 (constant 1024 of type byte) overflows byte
// line 57: cannot use 1000 (untyped int constant) as byte value in variable declaration (overflows)
// line 58: constant 1000 overflows byte
// line 59: byte(100) * byte(100) (constant 10000 of type byte) overflows byte
// line 60: byte(100) * 100 (constant 10000 of type byte) overflows byte
// line 61: 1000 (untyped int constant) overflows byte
// line 63: constant 1000 overflows byte
// line 64: invalid operation: division by zero
// line 67: Big * Big (constant 1e+600 of type float64) overflows float64
// line 68: float64(Big) * Big (constant 1e+600 of type float64) overflows float64
// line 69: Big * Big (constant 1e+600 of type float64) overflows float64
// line 70: invalid operation: division by zero
// line 71: invalid operation: operator % not defined on 1000 (untyped float constant)
// line 77: cannot use Int8 (constant 101 of type int8) as int value in argument to f
// line 78: cannot use Minus1 (constant -1 of type int8) as int value in argument to f
// line 79: cannot use Uint8 (constant 102 of type uint8) as int value in argument to f
// line 81: cannot use Float32 (constant 104.5 of type float32) as int value in argument to f
// line 82: cannot use Float64 (constant 105.5 of type float64) as int value in argument to f
// line 83: cannot use ConstFloat (untyped float constant 106.5) as int value in argument to f (truncated)
// line 85: cannot use Big (constant 1e+300 of type float64) as int value in argument to f
// line 86: cannot use String (untyped string constant "abc") as int value in argument to f
// line 87: cannot use Bool (untyped bool constant true) as int value in argument to f
// line 90: nil is not constant
// line 91: string([]byte(nil)) (value of type string) is not constant
// line 94: (*int)(nil) (value of type *int) is not constant
