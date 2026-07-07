// errorcheck

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that we don't get spurious follow-on errors
// after a missing expression. Specifically, the parser
// shouldn't skip over closing parentheses of any kind.

package p

func _() {
	go func() {     // no error here about goroutine
		send <- // GCCGO_ERROR "undefined name"
	}()             // ERROR "expected expression|expected operand"
}

func _() {
	defer func() { // no error here about deferred function
		1 +    // GCCGO_ERROR "value computed is not used"
	}()            // ERROR "expected expression|expected operand"
}

func _() {
	_ = (1 +)             // ERROR "expected expression|expected operand"
	_ = a[2 +]            // ERROR "expected expression|expected operand|undefined name"
	_ = []int{1, 2, 3 + } // ERROR "expected expression|expected operand"
}

// GnoError:
// line 16: expected operand, found '}' (and 5 more errors)
// line 17: expected operand, found '}' (and 5 more errors)
// line 19: expected '(', found _ (and 3 more errors)
// line 20: expected operand, found 'defer' (and 4 more errors)
// line 22: expected operand, found '}' (and 1 more errors)
// line 23: expected operand, found '}' (and 1 more errors)
// line 25: expected '(', found _ (and 3 more errors)
// line 26: expected '==', found '=' (and 1 more errors)
// line 27: expected '==', found '=' (and 1 more errors)
// line 28: expected '==', found '=' (and 1 more errors)
// line 29: expected operand, found '}' (and 1 more errors)
// line 31: expected '(', found main

// GoTypeCheckError:
// line 16: expected operand, found '}' (and 5 more errors)
// line 22: expected operand, found '}' (and 1 more errors)
// line 26: expected '==', found '=' (and 1 more errors)
// line 27: expected '==', found '=' (and 1 more errors)
// line 28: expected '==', found '=' (and 1 more errors)

// GnoOverStrictError:
// line 17: expected operand, found '}' (and 5 more errors)
// line 19: expected '(', found _ (and 3 more errors)
// line 20: expected operand, found 'defer' (and 4 more errors)
// line 23: expected operand, found '}' (and 1 more errors)
// line 25: expected '(', found _ (and 3 more errors)
// line 29: expected operand, found '}' (and 1 more errors)
// line 31: expected '(', found main
