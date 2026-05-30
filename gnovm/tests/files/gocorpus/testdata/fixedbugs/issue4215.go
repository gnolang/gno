// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func foo() (int, int) {
	return 2.3 // ERROR "not enough return values\n\thave \(number\)\n\twant \(int, int\)|not enough arguments to return"
}

func foo2() {
	return int(2), 2 // ERROR "too many (arguments to return|return values)\n\thave \(int, number\)\n\twant \(\)|return with value in function with no return type"
}

func foo3(v int) (a, b, c, d int) {
	if v >= 0 {
		return 1 // ERROR "not enough return values\n\thave \(number\)\n\twant \(int, int, int, int\)|not enough arguments to return"
	}
	return 2, 3 // ERROR "not enough return values\n\thave \(number, number\)\n\twant \(int, int, int, int\)|not enough arguments to return"
}

func foo4(name string) (string, int) {
	switch name {
	case "cow":
		return "moo" // ERROR "not enough return values\n\thave \(string\)\n\twant \(string, int\)|not enough arguments to return"
	case "dog":
		return "dog", 10, true // ERROR "too many return values\n\thave \(string, number, bool\)\n\twant \(string, int\)|too many arguments to return"
	case "fish":
		return "" // ERROR "not enough return values\n\thave \(string\)\n\twant \(string, int\)|not enough arguments to return"
	default:
		return "lizard", 10
	}
}

type S int
type T string
type U float64

func foo5() (S, T, U) {
	if false {
		return "" // ERROR "not enough return values\n\thave \(string\)\n\twant \(S, T, U\)|not enough arguments to return"
	} else {
		ptr := new(T)
		return ptr // ERROR "not enough return values\n\thave \(\*T\)\n\twant \(S, T, U\)|not enough arguments to return"
	}
	return new(S), 12.34, 1 + 0i, 'r', true // ERROR "too many return values\n\thave \(\*S, number, number, number, bool\)\n\twant \(S, T, U\)|too many arguments to return"
}

func foo6() (T, string) {
	return "T", true, true // ERROR "too many return values\n\thave \(string, bool, bool\)\n\twant \(T, string\)|too many arguments to return"
}

// GnoError:
// line 10: expected 2 return values

// GoTypeCheckError:
// line 10: not enough return values
// 	have (number)
// 	want (int, int)
// line 14: too many return values
// 	have (int, number)
// 	want ()
// line 19: not enough return values
// 	have (number)
// 	want (int, int, int, int)
// line 21: not enough return values
// 	have (number, number)
// 	want (int, int, int, int)
// line 27: not enough return values
// 	have (string)
// 	want (string, int)
// line 29: too many return values
// 	have (string, number, bool)
// 	want (string, int)
// line 31: not enough return values
// 	have (string)
// 	want (string, int)
// line 43: not enough return values
// 	have (string)
// 	want (S, T, U)
// line 46: not enough return values
// 	have (*T)
// 	want (S, T, U)
// line 48: too many return values
// 	have (*S, number, number, number, bool)
// 	want (S, T, U)
// line 52: too many return values
// 	have (string, bool, bool)
// 	want (T, string)
