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
// line 10: missing condition in if statement (and 5 more errors)
// line 13: missing condition in if statement (and 4 more errors)
// line 15: expected operand, found 'if' (and 4 more errors)
// line 17: expected operand, found 'if' (and 3 more errors)
// line 19: expected operand, found 'if' (and 2 more errors)
// line 20: missing condition in if statement (and 1 more errors)
// line 22: expected operand, found 'if' (and 1 more errors)
// line 24: expected operand, found 'if' (and 2 more errors)
// line 26: expected operand, found 'if' (and 1 more errors)
// line 27: missing condition in if statement
// line 28: expected operand, found '}' (and 1 more errors)
// line 30: expected '(', found main

// GoTypeCheckError:
// line 10: missing condition in if statement (and 5 more errors)
// line 13: missing condition in if statement (and 4 more errors)
// line 15: expected operand, found 'if' (and 4 more errors)
// line 17: expected operand, found 'if' (and 3 more errors)
// line 19: expected operand, found 'if' (and 2 more errors)
// line 26: expected operand, found 'if' (and 1 more errors)

// GnoOverStrictError:
// line 20: missing condition in if statement (and 1 more errors)
// line 22: expected operand, found 'if' (and 1 more errors)
// line 24: expected operand, found 'if' (and 2 more errors)
// line 27: missing condition in if statement
// line 28: expected operand, found '}' (and 1 more errors)
// line 30: expected '(', found main
