// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file contains test cases for various forms of
// method receiver declarations, per the spec clarification
// https://golang.org/cl/142757.

package issue28251

// test case from issue28251
type T struct{}

type T0 = *T

func (T0) m() {}

func _() { (&T{}).m() }

// various alternative forms
type (
        T1 = (((T)))
)

func ((*(T1))) m1() {}
func _() { (T{}).m2() }
func _() { (&T{}).m2() }

type (
        T2 = (((T3)))
        T3 = T
)

func (T2) m2() {}
func _() { (T{}).m2() }
func _() { (&T{}).m2() }

type (
        T4 = ((*(T5)))
        T5 = T
)

func (T4) m4() {}
func _() { (T{}).m4 /* ERROR "cannot call pointer method m4 on T" */ () }
func _() { (&T{}).m4() }

type (
        T6 = (((T7)))
        T7 = (*(T8))
        T8 = T
)

func (T6) m6() {}
func _() { (T{}).m6 /* ERROR "cannot call pointer method m6 on T" */ () }
func _() { (&T{}).m6() }

type (
        T9 = *T10
        T10 = *T11
        T11 = T
)

func (T9 /* ERRORx `invalid receiver type (\*\*T|T9)` */ ) m9() {}
func _() { (T{}).m9 /* ERROR "has no field or method m9" */ () }
func _() { (&T{}).m9 /* ERROR "has no field or method m9" */ () }

// GnoError:
// line 26: missing field m2 in gno.land/p/filetest/issue28251.T
// line 27: missing field m2 in *gno.land/p/filetest/issue28251.T
// line 30: StaticBlock.Define2(T2) cannot change .V
// line 35: missing field m2 in gno.land/p/filetest/issue28251.T
// line 36: missing field m2 in *gno.land/p/filetest/issue28251.T
// line 48: StaticBlock.Define2(T6) cannot change .V
// line 54: missing field m6 in gno.land/p/filetest/issue28251.T
// line 55: missing field m6 in *gno.land/p/filetest/issue28251.T
// line 63: invalid receiver type **gno.land/p/filetest/issue28251.T (base type is pointer type)
// line 64: missing field m9 in gno.land/p/filetest/issue28251.T
// line 65: missing field m9 in *gno.land/p/filetest/issue28251.T

// GoTypeCheckError:
// line 44
// line 54
// line 63
// line 64
// line 65

// GoTypesOnly:
// line 44: gc expects: cannot call pointer method m4 on T

// GnoOverStrictError:
// line 26: missing field m2 in gno.land/p/filetest/issue28251.T
// line 27: missing field m2 in *gno.land/p/filetest/issue28251.T
// line 30: StaticBlock.Define2(T2) cannot change .V
// line 35: missing field m2 in gno.land/p/filetest/issue28251.T
// line 36: missing field m2 in *gno.land/p/filetest/issue28251.T
// line 48: StaticBlock.Define2(T6) cannot change .V
// line 55: missing field m6 in *gno.land/p/filetest/issue28251.T

// KnownIssue:
// errorcheck: 7 over-strict line(s), 0 uncaught marker(s)
