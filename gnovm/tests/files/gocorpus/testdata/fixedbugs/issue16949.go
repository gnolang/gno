// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Ensure that typed non-integer len and cap make arguments are not accepted.

package main

var sink []byte

func main() {
	sink = make([]byte, 1.0)
	sink = make([]byte, float32(1.0)) // ERROR "non-integer.*len|must be integer"
	sink = make([]byte, float64(1.0)) // ERROR "non-integer.*len|must be integer"

	sink = make([]byte, 0, 1.0)
	sink = make([]byte, 0, float32(1.0)) // ERROR "non-integer.*cap|must be integer"
	sink = make([]byte, 0, float64(1.0)) // ERROR "non-integer.*cap|must be integer"

	sink = make([]byte, 1+0i)
	sink = make([]byte, complex64(1+0i))  // ERROR "non-integer.*len|must be integer"
	sink = make([]byte, complex128(1+0i)) // ERROR "non-integer.*len|must be integer"

	sink = make([]byte, 0, 1+0i)
	sink = make([]byte, 0, complex64(1+0i))  // ERROR "non-integer.*cap|must be integer"
	sink = make([]byte, 0, complex128(1+0i)) // ERROR "non-integer.*cap|must be integer"

}

// GnoError:
// line 15: invalid argument: index (const (1 float32)) (variable of type float32) must be integer
// line 16: invalid argument: index (const (1 float64)) (variable of type float64) must be integer
// line 19: invalid argument: index (const (1 float32)) (variable of type float32) must be integer
// line 20: invalid argument: index (const (1 float64)) (variable of type float64) must be integer

// GoTypeCheckError:
// line 15: invalid argument: index float32(1.0) (constant 1 of type float32) must be integer
// line 16: invalid argument: index float64(1.0) (constant 1 of type float64) must be integer
// line 19: invalid argument: index float32(1.0) (constant 1 of type float32) must be integer
// line 20: invalid argument: index float64(1.0) (constant 1 of type float64) must be integer
// line 23: invalid argument: index complex64(1 + 0i) (constant (1 + 0i) of type complex64) must be integer
// line 24: invalid argument: index complex128(1 + 0i) (constant (1 + 0i) of type complex128) must be integer
// line 27: invalid argument: index complex64(1 + 0i) (constant (1 + 0i) of type complex64) must be integer
// line 28: invalid argument: index complex128(1 + 0i) (constant (1 + 0i) of type complex128) must be integer
