// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tests for golang.org/issue/13471

package main

func main() {
	const _ int64 = 1e646456992 // ERROR "integer too large|floating-point constant truncated to integer|exponent too large|truncated"
	const _ int32 = 1e64645699  // ERROR "integer too large|floating-point constant truncated to integer|exponent too large|truncated"
	const _ int16 = 1e6464569   // ERROR "integer too large|floating-point constant truncated to integer|exponent too large|truncated"
	const _ int8 = 1e646456     // ERROR "integer too large|floating-point constant truncated to integer|exponent too large|truncated"
	const _ int = 1e64645       // ERROR "integer too large|floating-point constant truncated to integer|exponent too large|truncated"

	const _ uint64 = 1e646456992 // ERROR "integer too large|floating-point constant truncated to integer|exponent too large|truncated"
	const _ uint32 = 1e64645699  // ERROR "integer too large|floating-point constant truncated to integer|exponent too large|truncated"
	const _ uint16 = 1e6464569   // ERROR "integer too large|floating-point constant truncated to integer|exponent too large|truncated"
	const _ uint8 = 1e646456     // ERROR "integer too large|floating-point constant truncated to integer|exponent too large|truncated"
	const _ uint = 1e64645       // ERROR "integer too large|floating-point constant truncated to integer|exponent too large|truncated"

	const _ rune = 1e64645 // ERROR "integer too large|floating-point constant truncated to integer|exponent too large|truncated"
}

// GnoError:
// line 12: bigint overflows target kind
// line 13: bigint overflows target kind
// line 14: bigint overflows target kind
// line 15: bigint overflows target kind
// line 16: bigint overflows target kind
// line 18: bigint overflows target kind
// line 19: bigint overflows target kind
// line 20: bigint overflows target kind
// line 21: bigint overflows target kind
// line 22: bigint overflows target kind
// line 24: bigint overflows target kind

// GoTypeCheckError:
// line 12
// line 13
// line 14
// line 15
// line 16
// line 18
// line 19
// line 20
// line 21
// line 22
// line 24

// Verdict: Clean errorcheck
