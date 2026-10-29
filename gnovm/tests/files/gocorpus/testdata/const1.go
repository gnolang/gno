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

// GnoTrim: lines 12, 92-93

// GnoTrimmed: lines 92, 93

// GnoError:
// line 10: constant overflows
// line 35: bigint overflows target kind
// line 36: bigint overflows target kind
// line 37: bigint overflows target kind
// line 38: bigint overflows target kind
// line 44: bigint underflows target kind
// line 47: bigint underflows target kind
// line 51: constant -1 overflows Uint8Kind
// line 52: constant -1 overflows Uint8Kind
// line 53: bigint underflows target kind
// line 55: bigint overflows target kind
// line 57: bigint overflows target kind
// line 58: bigint overflows target kind
// line 61: bigint overflows target kind
// line 63: bigint overflows target kind
// line 64: invalid operation: division by zero
// line 70: invalid operation: division by zero
// line 71: operator % not defined on: BigdecKind
// line 74: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 77: cannot use int8 as int
// line 78: cannot use int8 as int
// line 79: cannot use uint8 as int
// line 81: cannot use float32 as int
// line 82: cannot use float64 as int
// line 83: cannot convert untyped bigdec to integer -- 106.5 not an exact integer
// line 85: cannot use float64 as int
// line 86: cannot use untyped string as IntKind
// line 87: cannot use untyped bool as IntKind
// line 90: use of untyped nil in variable declaration
// line 94: (const (nil *int)) (variable of type *int) is not constant

// Verdict: KnownIssue errorcheck: #6392 uncaught markers: lines 33, 39, 40, 43, 46, 59, 60, 67, 68, 69; #6344 uncaught markers: lines 56, 91; over-strict: lines 10 (the shift overflow of line 56, reported at the package line)
