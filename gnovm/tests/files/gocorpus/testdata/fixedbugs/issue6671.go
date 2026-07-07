// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 6671: Logical operators should produce untyped bool for untyped operands.

package p

type mybool bool

func _(x, y int) {
	type mybool bool
	var b mybool
	_ = b
	b = bool(true)             // ERROR "cannot use"
	b = true                   // permitted as expected
	b = bool(true) && true     // ERROR "cannot use"
	b = true && true           // permitted => && returns an untyped bool
	b = x < y                  // permitted => x < y returns an untyped bool
	b = true && x < y          // permitted => result of && returns untyped bool
	b = x < y && x < y         // permitted => result of && returns untyped bool
	b = x < y || x < y         // permitted => result of || returns untyped bool
	var c bool = true && x < y // permitted => result of && is bool
	c = false || x < y         // permitted => result of || returns untyped bool
	_ = c
}

// GnoError:
// line 17: cannot use bool as gno.land/p/filetest/p[gno.land/p/filetest/p/issue6671.go:14:1-29:2].mybool without explicit conversion
// line 19: cannot use bool as gno.land/p/filetest/p[gno.land/p/filetest/p/issue6671.go:14:1-29:2].mybool without explicit conversion
// line 23: cannot use bool as gno.land/p/filetest/p[gno.land/p/filetest/p/issue6671.go:14:1-29:2].mybool without explicit conversion
// line 24: cannot use bool as gno.land/p/filetest/p[gno.land/p/filetest/p/issue6671.go:14:1-29:2].mybool without explicit conversion

// GoTypeCheckError:
// line 17: cannot use bool(true) (constant true of type bool) as mybool value in assignment
// line 19: cannot use bool(true) && true (constant true of type bool) as mybool value in assignment

// GnoOverStrictError:
// line 23: cannot use bool as gno.land/p/filetest/p[gno.land/p/filetest/p/issue6671.go:14:1-29:2].mybool without explicit conversion
// line 24: cannot use bool as gno.land/p/filetest/p[gno.land/p/filetest/p/issue6671.go:14:1-29:2].mybool without explicit conversion
