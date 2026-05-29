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
// line 12: cannot use 1e646456992 (untyped float constant 1e+646456992) as int64 value in constant declaration (truncated)
// line 13: cannot use 1e64645699 (untyped float constant 1e+64645699) as int32 value in constant declaration (truncated)
// line 14: cannot use 1e6464569 (untyped float constant 1e+6464569) as int16 value in constant declaration (truncated)
// line 15: cannot use 1e646456 (untyped float constant 1e+646456) as int8 value in constant declaration (truncated)
// line 16: cannot use 1e64645 (untyped float constant 1e+64645) as int value in constant declaration (truncated)
// line 18: cannot use 1e646456992 (untyped float constant 1e+646456992) as uint64 value in constant declaration (truncated)
// line 19: cannot use 1e64645699 (untyped float constant 1e+64645699) as uint32 value in constant declaration (truncated)
// line 20: cannot use 1e6464569 (untyped float constant 1e+6464569) as uint16 value in constant declaration (truncated)
// line 21: cannot use 1e646456 (untyped float constant 1e+646456) as uint8 value in constant declaration (truncated)
// line 22: cannot use 1e64645 (untyped float constant 1e+64645) as uint value in constant declaration (truncated)
// line 24: cannot use 1e64645 (untyped float constant 1e+64645) as rune value in constant declaration (truncated)
