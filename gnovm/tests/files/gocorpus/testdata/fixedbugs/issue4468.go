// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 4468: go/defer calls may not be parenthesized.

package p

type T int

func (t *T) F() T {
	return *t
}

type S struct {
	t T
}

func F() {
	go F            // ERROR "must be function call"
	defer F         // ERROR "must be function call"
	go (F)		// ERROR "must be function call|must not be parenthesized"
	defer (F)	// ERROR "must be function call|must not be parenthesized"
	go (F())	// ERROR "must be function call|must not be parenthesized"
	defer (F())	// ERROR "must be function call|must not be parenthesized"
	var s S
	(&s.t).F()
	go (&s.t).F()
	defer (&s.t).F()
}

// GnoError:
// line 22: expression in go must be function call (and 5 more errors)
// line 23: expression in defer must be function call (and 4 more errors)
// line 24: expression in go must not be parenthesized (and 3 more errors)
// line 25: expression in defer must not be parenthesized (and 2 more errors)
// line 26: expression in go must not be parenthesized (and 1 more errors)
// line 27: expression in defer must not be parenthesized
// line 30: goroutines are not permitted

// GoTypeCheckError:
// line 22: expression in go must be function call (and 5 more errors)
// line 23: expression in defer must be function call (and 4 more errors)
// line 24: expression in go must not be parenthesized (and 3 more errors)
// line 25: expression in defer must not be parenthesized (and 2 more errors)
// line 26: expression in go must not be parenthesized (and 1 more errors)
// line 27: expression in defer must not be parenthesized

// GnoOverStrictError:
// line 30: goroutines are not permitted
