// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// https://code.google.com/p/gofrontend/issues/detail?id=1

package main

func f1() {
	a, b := f() // ERROR "assignment mismatch|does not match|cannot initialize"
	_, _ = a, b
}

func f2() {
	var a, b int
	a, b = f() // ERROR "assignment mismatch|does not match|cannot assign"
	_, _ = a, b
}

func f() int {
	return 1
}

// GnoError:
// line 12: assignment mismatch: 2 variables but f<VPBlock(3,2)> returns 1 values

// GoTypeCheckError:
// line 12: assignment mismatch: 2 variables but f returns 1 value
// line 18: assignment mismatch: 2 variables but f returns 1 value
