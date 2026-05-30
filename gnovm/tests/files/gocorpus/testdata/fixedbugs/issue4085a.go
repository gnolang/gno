// errorcheck

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T []int

func main() {
	_ = make(T, -1)    // ERROR "negative"
	_ = make(T, 0.5)   // ERROR "constant 0.5 truncated to integer|non-integer len argument|truncated to int"
	_ = make(T, 1.0)   // ok
	_ = make(T, 1<<63) // ERROR "len argument too large|overflows int"
	_ = make(T, 0, -1) // ERROR "negative cap|must not be negative"
	_ = make(T, 10, 0) // ERROR "len larger than cap|length and capacity swapped"
}

// GnoError:
// line 12: invalid argument: index (-1 <untyped> bigint) must not be negative
// line 13: cannot convert untyped bigdec to integer -- 0.5 not an exact integer
// line 15: bigint overflows target kind
// line 16: invalid argument: index (-1 <untyped> bigint) must not be negative
// line 17: invalid argument: len larger than cap in make(main.T)

// GoTypeCheckError:
// line 12: invalid argument: index -1 (constant of type int) must not be negative
// line 13: 0.5 (untyped float constant) truncated to int
// line 15: 1 << 63 (untyped int constant 9223372036854775808) overflows int
// line 16: invalid argument: index -1 (constant of type int) must not be negative
// line 17: invalid argument: length and capacity swapped
