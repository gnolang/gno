// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 1192 - detail in error

package main

func foo() (a, b, c int) {
	return 0, 1 2.01  // ERROR "unexpected literal 2.01|expected ';' or '}' or newline|not enough arguments to return"
}

// GnoError:
// line 11: 2: [function "foo" does not terminate]
// line 12: expected ';', found 2.01 (and 1 more errors)
// line 13: expected declaration, found '}'

// GoTypeCheckError:
// line 12: expected ';', found 2.01 (and 1 more errors)

// GnoOverStrictError:
// line 11: 2: [function "foo" does not terminate]
// line 13: expected declaration, found '}'
