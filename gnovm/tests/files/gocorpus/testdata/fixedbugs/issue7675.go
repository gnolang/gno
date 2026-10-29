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
// line 16: wrong argument count in call to f<VPBlock(3,0)>; want 4 got 3
// line 18: wrong argument count in call to f<VPBlock(3,0)>; want 4 got 5
// line 19: not enough arguments in call to g<VPBlock(3,1)>; want 3 (besides variadic) got 2
// line 21: not enough arguments in call to g<VPBlock(3,1)>; want 4 (including variadic) got 3
// line 23: not enough arguments in call to g<VPBlock(3,1)>; want 4 (including variadic) got 5

// GnoWordingDiffers:
// line 16: gc expects: not enough arguments|incompatible type; gno: wrong argument count in call to f<VPBlock(3,0)>; want 4 got 3
// line 18: gc expects: too many arguments|incompatible type; gno: wrong argument count in call to f<VPBlock(3,0)>; want 4 got 5
// line 23: gc expects: too many arguments|truncated to integer; gno: not enough arguments in call to g<VPBlock(3,1)>; want 4 (including variadic) got 5

// Verdict: Clean errorcheck
