// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that erroneous switch statements are detected by the compiler.
// Does not compile.

package main

func f() {
	switch {
	case 0; // ERROR "expecting := or = or : or comma|expected :"
	}

	switch {
	case 0; // ERROR "expecting := or = or : or comma|expected :"
	default:
	}

	switch {
	case 0: case 0: default:
	}

	switch {
	case 0: f(); case 0:
	case 0: f() case 0: // ERROR "unexpected keyword case at end of statement"
	}

	switch {
	case 0: f(); default:
	case 0: f() default: // ERROR "unexpected keyword default at end of statement"
	}

	switch {
	if x: // ERROR "expected case or default or }"
	}
}

// GnoError:
// line 14: expected ':', found ';'

// GoTypeCheckError:
// line 14

// ParseBlocked:
// line 18: gc expects: expecting := or = or : or comma|expected :
// line 28: gc expects: unexpected keyword case at end of statement
// line 33: gc expects: unexpected keyword default at end of statement
// line 37: gc expects: expected case or default or }

// GnoWordingDiffers:
// line 14: gc expects: expecting := or = or : or comma|expected :; gno: expected ':', found ';'

// Verdict: Clean errorcheck
