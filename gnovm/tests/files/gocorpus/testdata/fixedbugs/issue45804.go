// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func g() int
func h(int)

var b bool

func f() {
	did := g()
	if !did && b { // ERROR "invalid operation"
		h(x) // ERROR "undefined"
	}
}

// GnoError:
// line 9: function g does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 10: function h does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 16: name did not declared

// GoTypeCheckError:
// line 16
// line 17

// GoTypesOnly:
// line 17: gc expects: undefined

// GnoOverStrictError:
// line 9: function g does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 10: function h does not have a body but is not natively defined (did you build after pulling from the repository?)

// Verdict: KnownIssue errorcheck: 2 over-strict line(s), 0 uncaught marker(s)
