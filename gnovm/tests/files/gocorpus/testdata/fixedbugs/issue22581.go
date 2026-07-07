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
// line 13: expected declaration, found 'if'
// line 14: expected declaration, found '}'
// line 16: expected declaration, found 'switch'
// line 17: expected declaration, found '}'
// line 19: expected declaration, found 'switch'
// line 20: expected declaration, found '}'
// line 22: expected declaration, found 'for'
// line 23: expected declaration, found '}'
// line 25: expected declaration, found 'for'
// line 26: expected declaration, found '}'
// line 27: expected declaration, found '}'

// GoTypeCheckError:
// line 10: expected ';', found ')' (and 8 more errors)
// line 13: expected declaration, found 'if'
// line 16: expected declaration, found 'switch'
// line 19: expected declaration, found 'switch'
// line 22: expected declaration, found 'for'
// line 25: expected declaration, found 'for'

// GnoOverStrictError:
// line 14: expected declaration, found '}'
// line 17: expected declaration, found '}'
// line 20: expected declaration, found '}'
// line 23: expected declaration, found '}'
// line 26: expected declaration, found '}'
// line 27: expected declaration, found '}'
