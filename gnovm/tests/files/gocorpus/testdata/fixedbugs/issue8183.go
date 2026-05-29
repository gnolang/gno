// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tests correct reporting of line numbers for errors involving iota,
// Issue #8183.
package foo

const (
	ok = byte(iota + 253)
	bad
	barn
	bard // ERROR "constant 256 overflows byte|integer constant overflow|cannot convert"
)

const (
	c = len([1 - iota]int{})
	d
	e // ERROR "array bound must be non-negative|negative array bound|invalid array length"
	f // ERROR "array bound must be non-negative|negative array bound|invalid array length"
)

// GnoError:
// line 15: bigint overflows target kind

// GoTypeCheckError:
// line 21: invalid array length 1 - iota (untyped int constant -1)
// line 22: invalid array length 1 - iota (untyped int constant -2)
