// errorcheck -d=panic

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that illegal character literals are detected.
// Does not compile.

package main

const (
	// check that surrogate pair elements are invalid
	// (d800-dbff, dc00-dfff).
	_ = '\ud7ff' // ok
	_ = '\ud800'  // ERROR "Unicode|unicode"
	_ = "\U0000D999"  // ERROR "Unicode|unicode"
	_ = '\udc01' // ERROR "Unicode|unicode"
	_ = '\U0000dddd'  // ERROR "Unicode|unicode"
	_ = '\udfff' // ERROR "Unicode|unicode"
	_ = '\ue000' // ok
	_ = '\U0010ffff'  // ok
	_ = '\U00110000'  // ERROR "Unicode|unicode"
	_ = "abc\U0010ffffdef"  // ok
	_ = "abc\U00110000def"  // ERROR "Unicode|unicode"
	_ = '\Uffffffff'  // ERROR "Unicode|unicode"
)

// GnoError:
// line 16: escape sequence is invalid Unicode code point (and 7 more errors)
// line 17: escape sequence is invalid Unicode code point (and 6 more errors)
// line 18: escape sequence is invalid Unicode code point (and 5 more errors)
// line 19: escape sequence is invalid Unicode code point (and 4 more errors)
// line 20: escape sequence is invalid Unicode code point (and 3 more errors)
// line 23: escape sequence is invalid Unicode code point (and 2 more errors)
// line 25: escape sequence is invalid Unicode code point (and 1 more errors)
// line 26: escape sequence is invalid Unicode code point

// GoTypeCheckError:
// line 16: escape sequence is invalid Unicode code point (and 7 more errors)
// line 17: escape sequence is invalid Unicode code point (and 6 more errors)
// line 18: escape sequence is invalid Unicode code point (and 5 more errors)
// line 19: escape sequence is invalid Unicode code point (and 4 more errors)
// line 20: escape sequence is invalid Unicode code point (and 3 more errors)
// line 23: escape sequence is invalid Unicode code point (and 2 more errors)
// line 25: escape sequence is invalid Unicode code point (and 1 more errors)
// line 26: escape sequence is invalid Unicode code point
