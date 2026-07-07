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

// GnoOverStrictError:
// line 11: function f does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 16: not enough arguments in call to f
// 	have (number, number, string)
// 	want (string, int, float64, string)
// line 18: too many arguments in call to f
// 	have (number, number, string, number, number)
// 	want (string, int, float64, string)
// line 19: not enough arguments in call to g
// 	have (number, number)
// 	want (string, int, float64, ...string)
// line 21: not enough arguments in call to g
// 	have (number, number, []int...)
// 	want (string, int, float64, ...string)
// line 23: too many arguments in call to f
// 	have (number, number, string, number, number)
// 	want (string, int, float64, string)

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)
