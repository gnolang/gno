// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that erroneous switch statements are detected by the compiler.
// Does not compile.

package main

type I interface {
	M()
}

func bad() {
	var i I
	var s string

	switch i {
	case s: // ERROR "mismatched types string and I|incompatible types"
	}

	switch s {
	case i: // ERROR "mismatched types I and string|incompatible types"
	}

	var m, m1 map[int]int
	switch m {
	case nil:
	case m1: // ERROR "can only compare map m to nil|map can only be compared to nil|cannot compare"
	default:
	}

	var a, a1 []int
	switch a {
	case nil:
	case a1: // ERROR "can only compare slice a to nil|slice can only be compared to nil|cannot compare"
	default:
	}

	var f, f1 func()
	switch f {
	case nil:
	case f1: // ERROR "can only compare func f to nil|func can only be compared to nil|cannot compare"
	default:
	}

	var ar, ar1 [4]func()
	switch ar { // ERROR "cannot switch on"
	case ar1:
	default:
	}

	var st, st1 struct{ f func() }
	switch st { // ERROR "cannot switch on"
	case st1:
	}
}

func good() {
	var i interface{}
	var s string

	switch i {
	case s:
	}

	switch s {
	case i:
	}
}

// GnoError:
// line 21: string does not implement main.I (missing method M)
// line 25: cannot use main.I as string without explicit conversion
// line 29: 3: map[int]int can only be compared to nil
// line 36: 3: []int can only be compared to nil
// line 43: 3: func() can only be compared to nil
// line 50: 3: [4]func() is not comparable
// line 56: 3: struct{f func()} is not comparable
// line 70: cannot use interface {} as string

// GoTypeCheckError:
// line 21
// line 25
// line 31
// line 38
// line 45
// line 50
// line 56

// GnoOverStrictError:
// line 70: cannot use interface {} as string

// GnoWordingDiffers:
// line 21: gc expects: mismatched types string and I|incompatible types; gno: string does not implement main.I (missing method M)
// line 25: gc expects: mismatched types I and string|incompatible types; gno: cannot use main.I as string without explicit conversion
// line 31: gc expects: can only compare map m to nil|map can only be compared to nil|cannot compare; gno: 3: map[int]int can only be compared to nil
// line 38: gc expects: can only compare slice a to nil|slice can only be compared to nil|cannot compare; gno: 3: []int can only be compared to nil
// line 45: gc expects: can only compare func f to nil|func can only be compared to nil|cannot compare; gno: 3: func() can only be compared to nil
// line 50: gc expects: cannot switch on; gno: 3: [4]func() is not comparable
// line 56: gc expects: cannot switch on; gno: 3: struct{f func()} is not comparable

// Verdict: KnownIssue errorcheck: 1 over-strict line(s), 0 uncaught marker(s)
