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

// GoTypeCheckError:
// line 10
// line 11
// line 12
// line 15
// line 16
// line 17
// line 21
// line 22
// line 23
// line 25
// line 26
// line 27

// GoTypesOnly:
// line 11: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 12: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 15: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 16: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 17: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 21: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 22: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 23: gc expects: duplicate argument j|redefinition|previous|redeclared
// line 25: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 26: gc expects: duplicate argument i|redefinition|previous|redeclared
// line 27: gc expects: duplicate argument i|redefinition|previous|redeclared

// Verdict: Clean errorcheck
