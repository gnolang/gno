// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Ensure that typed non-integer, negative and too large
// values are not accepted as size argument in make for
// maps.

package main

type T map[int]int

var sink T

func main() {
	sink = make(T, -1)            // ERROR "negative size argument in make.*|must not be negative"
	sink = make(T, uint64(1<<63)) // ERROR "size argument too large in make.*|overflows int"

	// Test that errors are emitted at call sites, not const declarations
	const x = -1
	sink = make(T, x) // ERROR "negative size argument in make.*|must not be negative"
	const y = uint64(1 << 63)
	sink = make(T, y) // ERROR "size argument too large in make.*|overflows int"

	sink = make(T, 0.5) // ERROR "constant 0.5 truncated to integer|truncated to int"
	sink = make(T, 1.0)
	sink = make(T, float32(1.0)) // ERROR "non-integer size argument in make.*|must be integer"
	sink = make(T, float64(1.0)) // ERROR "non-integer size argument in make.*|must be integer"
	sink = make(T, 1+0i)
	sink = make(T, complex64(1+0i))  // ERROR "non-integer size argument in make.*|must be integer"
	sink = make(T, complex128(1+0i)) // ERROR "non-integer size argument in make.*|must be integer"
}

// GnoError:
// line 18: invalid argument: index (-1 <untyped> bigint) must not be negative
// line 23: invalid argument: index (-1 <untyped> bigint) must not be negative
// line 27: cannot convert untyped bigdec to integer -- 0.5 not an exact integer
// line 29: invalid argument: index (const (1 float32)) (variable of type float32) must be integer
// line 30: invalid argument: index (const (1 float64)) (variable of type float64) must be integer
// line 31: imaginaries are not supported
// line 32: name complex64 not declared
// line 33: name complex128 not declared

// GoTypeCheckError:
// line 18: invalid argument: index -1 (constant of type int) must not be negative
// line 19: invalid argument: index uint64(1 << 63) (constant 9223372036854775808 of type uint64) overflows int
// line 23: invalid argument: index x (constant -1 of type int) must not be negative
// line 25: invalid argument: index y (constant 9223372036854775808 of type uint64) overflows int
// line 27: 0.5 (untyped float constant) truncated to int
// line 29: invalid argument: index float32(1.0) (constant 1 of type float32) must be integer
// line 30: invalid argument: index float64(1.0) (constant 1 of type float64) must be integer
// line 32: invalid argument: index complex64(1 + 0i) (constant (1 + 0i) of type complex64) must be integer
// line 33: invalid argument: index complex128(1 + 0i) (constant (1 + 0i) of type complex128) must be integer

// GnoOverStrictError:
// line 31: imaginaries are not supported
