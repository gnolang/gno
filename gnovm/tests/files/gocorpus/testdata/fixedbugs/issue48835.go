// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f0()
func f1(_ int)
func f2(_, _ int)
func f2ddd(_, _ int, _ ...int)

func f() {
	var x int
	f0(1)              // ERROR "too many arguments in call to f0\n\thave \(number\)\n\twant \(\)"
	f0(x)              // ERROR "too many arguments in call to f0\n\thave \(int\)\n\twant \(\)"
	f1()               // ERROR "not enough arguments in call to f1\n\thave \(\)\n\twant \(int\)"
	f1(1, 2)           // ERROR "too many arguments in call to f1\n\thave \(number, number\)\n\twant \(int\)"
	f2(1)              // ERROR "not enough arguments in call to f2\n\thave \(number\)\n\twant \(int, int\)"
	f2(1, "foo", true) // ERROR "too many arguments in call to f2\n\thave \(number, string, bool\)\n\twant \(int, int\)"
	f2ddd(1)           // ERROR "not enough arguments in call to f2ddd\n\thave \(number\)\n\twant \(int, int, \.\.\.int\)"
	f2ddd(1, 2)
	f2ddd(1, 2, 3)
}

// GnoError:
// line 9: function f0 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 10: function f1 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 11: function f2 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 12: function f2ddd does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 16
// line 17
// line 18
// line 19
// line 20
// line 21
// line 22

// GoTypesOnly:
// line 16: gc expects an error (no pattern)
// line 17: gc expects an error (no pattern)
// line 18: gc expects an error (no pattern)
// line 19: gc expects an error (no pattern)
// line 20: gc expects an error (no pattern)
// line 21: gc expects an error (no pattern)
// line 22: gc expects an error (no pattern)

// GnoOverStrictError:
// line 9: function f0 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 10: function f1 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 11: function f2 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 12: function f2ddd does not have a body but is not natively defined (did you build after pulling from the repository?)

// Verdict: KnownIssue errorcheck: 4 over-strict line(s), 0 uncaught marker(s)
