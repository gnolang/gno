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

// GnoOverStrictError:
// line 11: function f does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 15: invalid case f in switch on x (func can only be compared to nil)

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)
