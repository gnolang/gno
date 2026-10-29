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
// line 18: expected ':', found ';'
// line 23: cannot use untyped Bigint as BoolKind
// line 27: cannot use untyped Bigint as BoolKind
// line 28: expected ';', found 'case'
// line 33: expected ';', found 'default'
// line 37: expected '}', found 'if'

// GnoWordingDiffers:
// line 14: gc expects: expecting := or = or : or comma|expected :; gno: expected ':', found ';'
// line 18: gc expects: expecting := or = or : or comma|expected :; gno: expected ':', found ';'
// line 28: gc expects: unexpected keyword case at end of statement; gno: expected ';', found 'case'
// line 33: gc expects: unexpected keyword default at end of statement; gno: expected ';', found 'default'
// line 37: gc expects: expected case or default or }; gno: expected '}', found 'if'

// Verdict: KnownIssue errorcheck: #6344 over-strict: lines 23, 27
