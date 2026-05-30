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

// GoTypeCheckError:
// line 21: invalid case s in switch on i (mismatched types string and I)
// line 25: invalid case i in switch on s (mismatched types I and string)
// line 31: invalid case m1 in switch on m (map can only be compared to nil)
// line 38: invalid case a1 in switch on a (slice can only be compared to nil)
// line 45: invalid case f1 in switch on f (func can only be compared to nil)
// line 50: cannot switch on ar (variable of type [4]func()) ([4]func() is not comparable)
// line 56: cannot switch on st (variable of type struct{f func()}) (struct{f func()} is not comparable)
