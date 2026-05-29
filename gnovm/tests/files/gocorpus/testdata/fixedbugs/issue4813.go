// errorcheck

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 4813: use of constant floats as indices.

package main

var A [3]int
var S []int
var T string

const (
	i  = 1
	f  = 2.0
	f2 = 2.1
	c  = complex(2, 0)
	c2 = complex(2, 1)
)

var (
	vf = f
	vc = c
)

var (
	a1 = A[i]
	a2 = A[f]
	a3 = A[f2] // ERROR "truncated|must be integer"
	a4 = A[c]
	a5 = A[c2] // ERROR "truncated|must be integer"
	a6 = A[vf] // ERROR "non-integer|must be integer"
	a7 = A[vc] // ERROR "non-integer|must be integer"

	s1 = S[i]
	s2 = S[f]
	s3 = S[f2] // ERROR "truncated|must be integer"
	s4 = S[c]
	s5 = S[c2] // ERROR "truncated|must be integer"
	s6 = S[vf] // ERROR "non-integer|must be integer"
	s7 = S[vc] // ERROR "non-integer|must be integer"

	t1 = T[i]
	t2 = T[f]
	t3 = T[f2] // ERROR "truncated|must be integer"
	t4 = T[c]
	t5 = T[c2] // ERROR "truncated|must be integer"
	t6 = T[vf] // ERROR "non-integer|must be integer"
	t7 = T[vc] // ERROR "non-integer|must be integer"
)

// GnoError:
// line 19: name complex not defined in fileset with files [issue4813.go]

// GoTypeCheckError:
// line 31: f2 (untyped float constant 2.1) truncated to int
// line 33: c2 (untyped complex constant (2 + 1i)) truncated to int
// line 34: invalid argument: index vf (variable of type float64) must be integer
// line 35: invalid argument: index vc (variable of type complex128) must be integer
// line 39: f2 (untyped float constant 2.1) truncated to int
// line 41: c2 (untyped complex constant (2 + 1i)) truncated to int
// line 42: invalid argument: index vf (variable of type float64) must be integer
// line 43: invalid argument: index vc (variable of type complex128) must be integer
// line 47: f2 (untyped float constant 2.1) truncated to int
// line 49: c2 (untyped complex constant (2 + 1i)) truncated to int
// line 50: invalid argument: index vf (variable of type float64) must be integer
// line 51: invalid argument: index vc (variable of type complex128) must be integer
