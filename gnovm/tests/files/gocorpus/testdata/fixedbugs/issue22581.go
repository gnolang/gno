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
// line 10: expected ';', found ')' (and 8 more errors)

// GoTypeCheckError:
// line 10

// UncaughtError:
// line 13: uncaught; gc expects an error (no pattern)
// line 16: uncaught; gc expects an error (no pattern)
// line 19: uncaught; gc expects an error (no pattern)
// line 22: uncaught; gc expects an error (no pattern)
// line 25: uncaught; gc expects an error (no pattern)

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 5 uncaught marker(s)
