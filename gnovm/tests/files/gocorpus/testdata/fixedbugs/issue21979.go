// errorcheck

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f() {
	_ = bool("")      // ERROR "cannot convert .. \(.*untyped string.*\) to type bool|invalid type conversion"
	_ = bool(1)       // ERROR "cannot convert 1 \(.*untyped int.*\) to type bool|invalid type conversion"
	_ = bool(1.0)     // ERROR "cannot convert 1.* \(.*untyped float.*\) to type bool|invalid type conversion"
	_ = bool(-4 + 2i) // ERROR "cannot convert -4 \+ 2i \(.*untyped complex.*\) to type bool|invalid type conversion"

	_ = string(true) // ERROR "cannot convert true \(.*untyped bool.*\) to type string|invalid type conversion"
	_ = string(-1)
	_ = string(1.0)     // ERROR "cannot convert 1.* \(.*untyped float.*\) to type string|invalid type conversion"
	_ = string(-4 + 2i) // ERROR "cannot convert -4 \+ 2i \(.*untyped complex.*\) to type string|invalid type conversion"

	_ = int("")   // ERROR "cannot convert .. \(.*untyped string.*\) to type int|invalid type conversion"
	_ = int(true) // ERROR "cannot convert true \(.*untyped bool.*\) to type int|invalid type conversion"
	_ = int(-1)
	_ = int(1)
	_ = int(1.0)
	_ = int(-4 + 2i) // ERROR "truncated to integer|cannot convert -4 \+ 2i \(.*untyped complex.*\) to type int"

	_ = uint("")   // ERROR "cannot convert .. \(.*untyped string.*\) to type uint|invalid type conversion"
	_ = uint(true) // ERROR "cannot convert true \(.*untyped bool.*\) to type uint|invalid type conversion"
	_ = uint(-1)   // ERROR "constant -1 overflows uint|integer constant overflow|cannot convert -1 \(untyped int constant\) to type uint"
	_ = uint(1)
	_ = uint(1.0)
	// types1 reports extra error "truncated to integer"
	_ = uint(-4 + 2i) // ERROR "constant -4 overflows uint|truncated to integer|cannot convert -4 \+ 2i \(untyped complex constant.*\) to type uint"

	_ = float64("")   // ERROR "cannot convert .. \(.*untyped string.*\) to type float64|invalid type conversion"
	_ = float64(true) // ERROR "cannot convert true \(.*untyped bool.*\) to type float64|invalid type conversion"
	_ = float64(-1)
	_ = float64(1)
	_ = float64(1.0)
	_ = float64(-4 + 2i) // ERROR "truncated to|cannot convert -4 \+ 2i \(.*untyped complex.*\) to type float64"

	_ = complex128("")   // ERROR "cannot convert .. \(.*untyped string.*\) to type complex128|invalid type conversion"
	_ = complex128(true) // ERROR "cannot convert true \(.*untyped bool.*\) to type complex128|invalid type conversion"
	_ = complex128(-1)
	_ = complex128(1)
	_ = complex128(1.0)
}

// GnoError:
// line 10: cannot convert StringKind to BoolKind
// line 11: cannot convert IntKind to BoolKind
// line 12: cannot convert Float64Kind to BoolKind
// line 13: imaginaries are not supported
// line 15: cannot convert BoolKind to StringKind
// line 17: cannot convert Float64Kind to StringKind
// line 18: imaginaries are not supported
// line 20: cannot convert StringKind to IntKind
// line 21: cannot convert BoolKind to IntKind
// line 25: imaginaries are not supported
// line 27: cannot convert StringKind to UintKind
// line 28: cannot convert BoolKind to UintKind
// line 29: bigint underflows target kind
// line 33: imaginaries are not supported
// line 35: cannot convert StringKind to Float64Kind
// line 36: cannot convert BoolKind to Float64Kind
// line 40: imaginaries are not supported
// line 42: name complex128 not declared
// line 43: name complex128 not declared
// line 44: name complex128 not declared
// line 45: name complex128 not declared
// line 46: name complex128 not declared

// GoTypeCheckError:
// line 10: cannot convert "" (untyped string constant) to type bool
// line 11: cannot convert 1 (untyped int constant) to type bool
// line 12: cannot convert 1.0 (untyped float constant 1) to type bool
// line 13: cannot convert -4 + 2i (untyped complex constant (-4 + 2i)) to type bool
// line 15: cannot convert true (untyped bool constant) to type string
// line 17: cannot convert 1.0 (untyped float constant 1) to type string
// line 18: cannot convert -4 + 2i (untyped complex constant (-4 + 2i)) to type string
// line 20: cannot convert "" (untyped string constant) to type int
// line 21: cannot convert true (untyped bool constant) to type int
// line 25: cannot convert -4 + 2i (untyped complex constant (-4 + 2i)) to type int
// line 27: cannot convert "" (untyped string constant) to type uint
// line 28: cannot convert true (untyped bool constant) to type uint
// line 29: constant -1 overflows uint
// line 33: cannot convert -4 + 2i (untyped complex constant (-4 + 2i)) to type uint
// line 35: cannot convert "" (untyped string constant) to type float64
// line 36: cannot convert true (untyped bool constant) to type float64
// line 40: cannot convert -4 + 2i (untyped complex constant (-4 + 2i)) to type float64
// line 42: cannot convert "" (untyped string constant) to type complex128
// line 43: cannot convert true (untyped bool constant) to type complex128

// GnoOverStrictError:
// line 44: name complex128 not declared
// line 45: name complex128 not declared
// line 46: name complex128 not declared
