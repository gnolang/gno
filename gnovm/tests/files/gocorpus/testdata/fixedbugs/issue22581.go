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

// GnoIncomplete: covered 2 of 6 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// GnoError:
// line 10: expected ';', found ')' (and 8 more errors)
// line 13: expected declaration, found 'if'

// GoTypeCheckError:
// line 10: expected ';', found ')' (and 8 more errors)
// line 13: expected declaration, found 'if'
