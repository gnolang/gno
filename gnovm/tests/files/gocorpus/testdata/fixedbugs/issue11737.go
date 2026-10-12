// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 11737 - invalid == not being caught until generated switch code was compiled

package p

func f()

func s(x interface{}) {
	switch x {
	case f: // ERROR "invalid case f \(type func\(\)\) in switch \(incomparable type\)|can only be compared to nil"
	}
}

// GnoError:
// line 11: function f does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 15

// GoTypesOnly:
// line 15: gc expects: invalid case f \(type func\(\)\) in switch \(incomparable type\)|can only be compared to nil

// GnoOverStrictError:
// line 11: function f does not have a body but is not natively defined (did you build after pulling from the repository?)

// Verdict: KnownIssue errorcheck: 1 over-strict line(s), 0 uncaught marker(s)
