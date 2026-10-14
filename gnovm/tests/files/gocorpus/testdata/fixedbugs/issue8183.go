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

// GoTypeCheckError:
// line 15
// line 21
// line 22

// GoTypesOnly:
// line 15: gc expects: constant 256 overflows byte|integer constant overflow|cannot convert
// line 21: gc expects: array bound must be non-negative|negative array bound|invalid array length
// line 22: gc expects: array bound must be non-negative|negative array bound|invalid array length

// Verdict: Clean errorcheck
