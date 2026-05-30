// errorcheck

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T interface {
	F1(i int) (i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"
	F2(i, i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"
	F3() (i, i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"
}

type T1 func(i, i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"
type T2 func(i int) (i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"
type T3 func() (i, i int) // ERROR "duplicate argument i|redefinition|previous|redeclared"

type R struct{}

func (i *R) F1(i int)         {} // ERROR "duplicate argument i|redefinition|previous|redeclared"
func (i *R) F2() (i int)      {return 0} // ERROR "duplicate argument i|redefinition|previous|redeclared"
func (i *R) F3(j int) (j int) {return 0} // ERROR "duplicate argument j|redefinition|previous|redeclared"

func F1(i, i int)      {} // ERROR "duplicate argument i|redefinition|previous|redeclared"
func F2(i int) (i int) {return 0} // ERROR "duplicate argument i|redefinition|previous|redeclared"
func F3() (i, i int)   {return 0, 0} // ERROR "duplicate argument i|redefinition|previous|redeclared"

// GnoError:
// line 10: i redeclared in this block
// 	previous declaration at funcdup.go:11:5 (and 10 more errors)
// line 11: i redeclared in this block
// 	previous declaration at funcdup.go:12:5 (and 10 more errors)
// line 12: i redeclared in this block
// 	previous declaration at funcdup.go:13:8 (and 9 more errors)
// line 15: i redeclared in this block
// 	previous declaration at funcdup.go:16:14 (and 8 more errors)
// line 16: i redeclared in this block
// 	previous declaration at funcdup.go:17:14 (and 7 more errors)
// line 17: i redeclared in this block
// 	previous declaration at funcdup.go:18:17 (and 6 more errors)
// line 21: i redeclared in this block
// 	previous declaration at funcdup.go:22:7 (and 5 more errors)
// line 22: i redeclared in this block
// 	previous declaration at funcdup.go:23:7 (and 4 more errors)
// line 23: j redeclared in this block
// 	previous declaration at funcdup.go:24:16 (and 3 more errors)
// line 25: i redeclared in this block
// 	previous declaration at funcdup.go:26:9 (and 2 more errors)
// line 26: i redeclared in this block
// 	previous declaration at funcdup.go:27:9 (and 1 more errors)
// line 27: i redeclared in this block
// 	previous declaration at funcdup.go:28:12

// GoTypeCheckError:
// line 10: i redeclared in this block
// line 11: i redeclared in this block
// line 12: i redeclared in this block
// line 15: i redeclared in this block
// line 16: i redeclared in this block
// line 17: i redeclared in this block
// line 21: i redeclared in this block
// line 22: i redeclared in this block
// line 23: j redeclared in this block
// line 25: i redeclared in this block
// line 26: i redeclared in this block
// line 27: i redeclared in this block
