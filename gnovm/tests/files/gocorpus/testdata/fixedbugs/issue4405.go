// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

const (
	_ = iota
	_ // ERROR "illegal character|invalid character"
	_  // ERROR "illegal character|invalid character"
	_  // ERROR "illegal character|invalid character"
	_  // ERROR "illegal character|invalid character"
)

// GnoError:
// line 11: illegal character U+0007 (and 4 more errors)
// line 12: illegal character U+0008 (and 3 more errors)
// line 13: illegal character U+000B (and 2 more errors)
// line 14: illegal character U+000C (and 1 more errors)

// GoTypeCheckError:
// line 11: illegal character U+0007 (and 4 more errors)
// line 12: illegal character U+0008 (and 3 more errors)
// line 13: illegal character U+000B (and 2 more errors)
// line 14: illegal character U+000C (and 1 more errors)
