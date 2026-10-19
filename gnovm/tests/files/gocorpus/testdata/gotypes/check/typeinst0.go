// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type myInt int

// Parameterized type declarations

// For now, a lone type parameter is not permitted as RHS in a type declaration (issue #45639).
type T1[P any] P // ERROR "cannot use a type parameter as RHS in type declaration"

type T2[P any] struct {
        f P
        g int // int should still be in scope chain
}

type List[P any] []P

// Pending clarification of #46477 we disallow aliases
// of generic types.
type A2 = List // ERROR "cannot use generic type"
var _ A2[int]
var _ A2

type A3 = List[int]
var _ A3

// Parameterized type instantiations

var x int
type _ x /* ERROR "not a type" */ [int]

type _ int /* ERROR "not a generic type" */ [] // ERROR "expected type argument list"
type _ myInt /* ERROR "not a generic type" */ [] // ERROR "expected type argument list"

// TODO(gri) better error messages
type _ T1[] // ERROR "expected type argument list"
type _ T1[x /* ERROR "not a type" */ ]
type _ T1 /* ERROR "too many type arguments for type T1: have 2, want 1" */ [int, float32]

var _ T2[int] = T2[int]{}

var _ List[int] = []int{1, 2, 3}
var _ List[[]int] = [][]int{{1, 2, 3}}
var _ List[List[List[int]]]

// Parameterized types containing parameterized types

type T3[P any] List[P]

var _ T3[int] = T3[int](List[int]{1, 2, 3})

// Self-recursive generic types are not permitted

type self1[P any] self1 /* ERROR "invalid recursive type" */ [P]
type self2[P any] *self2[P] // this is ok

// GnoError:
// line 12: name P not defined in fileset with files [typeinst0.go]
// line 15: name P not defined in fileset with files [typeinst0.go]
// line 19: name P not defined in fileset with files [typeinst0.go]
// line 23: should not happen
// line 27: unexpected type declaration type *gnolang.IndexExpr
// line 33: name not defined: x
// line 35: expected type argument list
// line 36: expected type argument list
// line 39: expected type argument list
// line 40: name not defined: x
// line 41: invalid operation: more than one index
// line 43: invalid operation: cannot index typeval{gno.land/p/filetest/p.T2} (value of type type{})
// line 45: invalid operation: cannot index typeval{gno.land/p/filetest/p.List} (value of type type{})
// line 46: invalid operation: cannot index typeval{gno.land/p/filetest/p.List} (value of type type{})
// line 47: invalid operation: cannot index typeval{gno.land/p/filetest/p.List} (value of type type{})
// line 51: unexpected type declaration type *gnolang.IndexExpr
// line 57: unexpected type declaration type *gnolang.IndexExpr
// line 58: name P not defined in fileset with files [typeinst0.go]

// GnoWordingDiffers:
// line 12: gc expects: cannot use a type parameter as RHS in type declaration; gno: name P not defined in fileset with files [typeinst0.go]
// line 33: gc expects: not a type; gno: name not defined: x
// line 40: gc expects: not a type; gno: name not defined: x
// line 41: gc expects: too many type arguments for type T1: have 2, want 1; gno: invalid operation: more than one index
// line 57: gc expects: invalid recursive type; gno: unexpected type declaration type *gnolang.IndexExpr

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 23; over-strict: lines 15, 19, 27, 43, 45, 46, 47, 51, 58; internal errors: lines 23
