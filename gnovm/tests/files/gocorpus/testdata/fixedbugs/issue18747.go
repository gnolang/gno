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

// GnoStaticIncomplete: covered 5 of 6 markers (Gno preprocess: 5, go/types guard: 5); Gno bailed before the rest — a runnable variant may exercise more

// GnoError:
// line 10: missing condition in if statement (and 5 more errors)
// line 13: missing condition in if statement (and 4 more errors)
// line 15: expected operand, found 'if' (and 4 more errors)
// line 17: expected operand, found 'if' (and 3 more errors)
// line 19: expected operand, found 'if' (and 2 more errors)

// GoTypeCheckError:
// line 10: missing condition in if statement (and 5 more errors)
// line 13: missing condition in if statement (and 4 more errors)
// line 15: expected operand, found 'if' (and 4 more errors)
// line 17: expected operand, found 'if' (and 3 more errors)
// line 19: expected operand, found 'if' (and 2 more errors)
