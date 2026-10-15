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
// 	previous declaration at funcdup.go:11:5
// line 11: i redeclared in this block
// 	previous declaration at funcdup.go:12:5
// line 12: i redeclared in this block
// 	previous declaration at funcdup.go:13:8
// line 15: i redeclared in this block
// 	previous declaration at funcdup.go:16:14
// line 16: i redeclared in this block
// 	previous declaration at funcdup.go:17:14
// line 17: i redeclared in this block
// 	previous declaration at funcdup.go:18:17
// line 21: i redeclared in this block
// 	previous declaration at funcdup.go:22:7
// line 22: i redeclared in this block
// 	previous declaration at funcdup.go:23:7
// line 23: j redeclared in this block
// 	previous declaration at funcdup.go:24:16
// line 25: i redeclared in this block
// 	previous declaration at funcdup.go:26:9
// line 26: i redeclared in this block
// 	previous declaration at funcdup.go:27:9

// ParseBlocked:
// line 27: gc expects: duplicate argument i|redefinition|previous|redeclared

// Verdict: Clean errorcheck
