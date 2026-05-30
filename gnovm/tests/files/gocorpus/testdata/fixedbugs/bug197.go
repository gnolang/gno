// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T int
type U int

var x int

var t T = int(0)	// ERROR "cannot use|incompatible"
var t1 T = int(x)	// ERROR "cannot use|incompatible"
var u U = int(0)	// ERROR "cannot use|incompatible"
var u1 U = int(x)	// ERROR "cannot use|incompatible"

type S string
var s S

var s1 = s + "hello"
var s2 = "hello" + s
var s3 = s + string("hello")	// ERROR "invalid operation|incompatible"
var s4 = string("hello") + s	// ERROR "invalid operation|incompatible"

var r string

var r1 = r + "hello"
var r2 = "hello" + r
var r3 = r + string("hello")
var r4 = string("hello") + r

// GnoError:
// line 14: cannot use int as main.T without explicit conversion
// line 15: cannot use int as main.T without explicit conversion
// line 16: cannot use int as main.U without explicit conversion
// line 17: cannot use int as main.U without explicit conversion
// line 24: invalid operation: s<VPBlock(2,4)> + (const ("hello" string)) (mismatched types main.S and string)
// line 25: invalid operation: (const ("hello" string)) + s<VPBlock(2,4)> (mismatched types string and main.S)

// GoTypeCheckError:
// line 14: cannot use int(0) (constant 0 of type int) as T value in variable declaration
// line 15: cannot use int(x) (value of type int) as T value in variable declaration
// line 16: cannot use int(0) (constant 0 of type int) as U value in variable declaration
// line 17: cannot use int(x) (value of type int) as U value in variable declaration
// line 24: invalid operation: s + string("hello") (mismatched types S and string)
// line 25: invalid operation: string("hello") + s (mismatched types string and S)
