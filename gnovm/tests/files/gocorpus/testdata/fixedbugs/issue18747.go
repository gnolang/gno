// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _ () {
	if {} // ERROR "missing condition in if statement"

	if
	{} // ERROR "missing condition in if statement"

	if ; {} // ERROR "missing condition in if statement"

	if foo; {} // ERROR "missing condition in if statement"

	if foo; // ERROR "missing condition in if statement"
	{}

	if foo {}

	if ; foo {}

	if foo // ERROR "unexpected newline, expected { after if clause"
	{}
}

// GnoError:
// line 10: missing condition in if statement

// ParseBlocked:
// line 13: gc expects: missing condition in if statement
// line 15: gc expects: missing condition in if statement
// line 17: gc expects: missing condition in if statement
// line 19: gc expects: missing condition in if statement
// line 26: gc expects: unexpected newline, expected { after if clause

// Verdict: Clean errorcheck
