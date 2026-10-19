// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f() {
	if i := g()); i == j { // ERROR "unexpected \)"
	}

	if i == g()] { // ERROR "unexpected \]"
	}

	switch i := g()); i { // ERROR "unexpected \)"
	}

	switch g()] { // ERROR "unexpected \]"
	}

	for i := g()); i < y; { // ERROR "unexpected \)"
	}

	for g()] { // ERROR "unexpected \]"
	}
}

// GnoError:
// line 10: expected ';', found ')'
// line 13: expected '{', found 'if'
// line 16: expected '{', found ')'
// line 19: expected '{', found ']'
// line 22: expected boolean or range expression, found assignment (missing parentheses around composite literal?)
// line 25: expected '{', found ']'
// line 27: expected '(', found main

// GnoWordingDiffers:
// line 10: gc expects: unexpected \); gno: expected ';', found ')'
// line 13: gc expects: unexpected \]; gno: expected '{', found 'if'
// line 16: gc expects: unexpected \); gno: expected '{', found ')'
// line 19: gc expects: unexpected \]; gno: expected '{', found ']'
// line 22: gc expects: unexpected \); gno: expected boolean or range expression, found assignment (missing parentheses around composite literal?)
// line 25: gc expects: unexpected \]; gno: expected '{', found ']'

// Verdict: KnownIssue errorcheck: #6344 over-strict: lines 27
