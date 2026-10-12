// errorcheck

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var T interface {
	F1(i int) (i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"
	F2(i, i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"
	F3() (i, i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"
}

var T1 func(i, i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"
var T2 func(i int) (i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"
var T3 func() (i, i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"

// GnoError:
// line 10: i redeclared in this block
// 	previous declaration at funcdup2.go:11:5 (and 5 more errors)

// GoTypeCheckError:
// line 10
// line 11
// line 12
// line 15
// line 16
// line 17

// GoTypesOnly:
// line 11: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 12: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 15: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 16: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 17: gc expects: duplicate argument i|redefinition|previous|redeclared

// Verdict: Clean errorcheck
