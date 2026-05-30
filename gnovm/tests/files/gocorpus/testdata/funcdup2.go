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
// line 11: i redeclared in this block
// 	previous declaration at funcdup2.go:12:5 (and 4 more errors)
// line 12: i redeclared in this block
// 	previous declaration at funcdup2.go:13:8 (and 3 more errors)
// line 15: i redeclared in this block
// 	previous declaration at funcdup2.go:16:13 (and 2 more errors)
// line 16: i redeclared in this block
// 	previous declaration at funcdup2.go:17:13 (and 1 more errors)
// line 17: i redeclared in this block
// 	previous declaration at funcdup2.go:18:16

// GoTypeCheckError:
// line 10: i redeclared in this block
// line 11: i redeclared in this block
// line 12: i redeclared in this block
// line 15: i redeclared in this block
// line 16: i redeclared in this block
// line 17: i redeclared in this block
