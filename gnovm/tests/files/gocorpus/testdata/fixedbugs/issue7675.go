// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 7675: fewer errors for wrong argument count

package p

func f(string, int, float64, string)

func g(string, int, float64, ...string)

func main() {
	f(1, 0.5, "hello") // ERROR "not enough arguments|incompatible type"
	f("1", 2, 3.1, "4")
	f(1, 0.5, "hello", 4, 5) // ERROR "too many arguments|incompatible type"
	g(1, 0.5)                // ERROR "not enough arguments|incompatible type"
	g("1", 2, 3.1)
	g(1, 0.5, []int{3, 4}...) // ERROR "not enough arguments|incompatible type"
	g("1", 2, 3.1, "4", "5")
	g(1, 0.5, "hello", 4, []int{5, 6}...) // ERROR "too many arguments|truncated to integer"
}

// GnoError:
// line 11: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 13: function g does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 16
// line 18
// line 19
// line 21
// line 23

// GoTypesOnly:
// line 16: gc expects: not enough arguments|incompatible type
// line 18: gc expects: too many arguments|incompatible type
// line 19: gc expects: not enough arguments|incompatible type
// line 21: gc expects: not enough arguments|incompatible type
// line 23: gc expects: too many arguments|truncated to integer

// GnoOverStrictError:
// line 11: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 13: function g does not have a body but is not natively defined (did you build after pulling from the repository?)

// Verdict: KnownIssue errorcheck: 2 over-strict line(s), 0 uncaught marker(s)
