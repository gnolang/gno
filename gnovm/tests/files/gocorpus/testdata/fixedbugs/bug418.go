// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 3044.
// Multiple valued expressions in return lists.

package p

func Two() (a, b int)

// F used to compile.
func F() (x interface{}, y int) {
	return Two(), 0 // ERROR "single-value context|2\-valued"
}

// Recursive used to trigger an internal compiler error.
func Recursive() (x interface{}, y int) {
	return Recursive(), 0 // ERROR "single-value context|2\-valued"
}

// GnoError:
// line 12: function Two does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 16
// line 21

// GoTypesOnly:
// line 16: gc expects an error (no pattern)
// line 21: gc expects an error (no pattern)

// GnoOverStrictError:
// line 12: function Two does not have a body but is not natively defined (did you build after pulling from the repository?)

// Verdict: KnownIssue errorcheck: 1 over-strict line(s), 0 uncaught marker(s)
