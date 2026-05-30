// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T struct {
	// legal according to spec
	x int
	y (int)
	int
	*float64
	// not legal according to spec
	(complex128)  // ERROR "non-declaration|expected|parenthesize"
	(*string)  // ERROR "non-declaration|expected|parenthesize"
	*(bool)    // ERROR "non-declaration|expected|parenthesize"
}

// legal according to spec
func (p T) m() {}

// now legal according to spec
func (p (T)) f() {}
func (p *(T)) g() {}
func (p (*T)) h() {}
func (p (*(T))) i() {}
func ((T),) j() {}

// GnoError:
// line 16: cannot parenthesize embedded type (and 2 more errors)
// line 17: cannot parenthesize embedded type (and 1 more errors)
// line 18: cannot parenthesize embedded type

// GoTypeCheckError:
// line 16: cannot parenthesize embedded type (and 2 more errors)
// line 17: cannot parenthesize embedded type (and 1 more errors)
// line 18: cannot parenthesize embedded type
