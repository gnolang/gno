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

// GoTypeCheckError:
// line 22

// UncaughtError:
// line 23: uncaught; gc expects: must be function call
// line 24: uncaught; gc expects: must be function call|must not be parenthesized
// line 25: uncaught; gc expects: must be function call|must not be parenthesized
// line 26: uncaught; gc expects: must be function call|must not be parenthesized
// line 27: uncaught; gc expects: must be function call|must not be parenthesized

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 5 uncaught marker(s)
