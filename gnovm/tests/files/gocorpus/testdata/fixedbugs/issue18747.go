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

// GoTypeCheckError:
// line 10

// UncaughtError:
// line 13: uncaught; gc expects: missing condition in if statement
// line 15: uncaught; gc expects: missing condition in if statement
// line 17: uncaught; gc expects: missing condition in if statement
// line 19: uncaught; gc expects: missing condition in if statement
// line 26: uncaught; gc expects: unexpected newline, expected { after if clause

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 5 uncaught marker(s)
