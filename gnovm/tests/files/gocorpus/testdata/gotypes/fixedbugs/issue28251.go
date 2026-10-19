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
// line 26: gno.land/p/filetest/issue28251.T{}.m2 undefined (type gno.land/p/filetest/issue28251.T has no field or method m2)
// line 27: (&gno.land/p/filetest/issue28251.T{}).m2 undefined (type *gno.land/p/filetest/issue28251.T has no field or method m2)
// line 30: StaticBlock.Define2(T2) cannot change .V
// line 35: gno.land/p/filetest/issue28251.T{}.m2 undefined (type gno.land/p/filetest/issue28251.T has no field or method m2)
// line 36: (&gno.land/p/filetest/issue28251.T{}).m2 undefined (type *gno.land/p/filetest/issue28251.T has no field or method m2)
// line 48: StaticBlock.Define2(T6) cannot change .V
// line 54: gno.land/p/filetest/issue28251.T{}.m6 undefined (type gno.land/p/filetest/issue28251.T has no field or method m6)
// line 55: (&gno.land/p/filetest/issue28251.T{}).m6 undefined (type *gno.land/p/filetest/issue28251.T has no field or method m6)
// line 63: invalid receiver type **gno.land/p/filetest/issue28251.T (base type is pointer type)
// line 64: gno.land/p/filetest/issue28251.T{}.m9 undefined (type gno.land/p/filetest/issue28251.T has no field or method m9)
// line 65: (&gno.land/p/filetest/issue28251.T{}).m9 undefined (type *gno.land/p/filetest/issue28251.T has no field or method m9)

// GnoWordingDiffers:
// line 54: gc expects: cannot call pointer method m6 on T; gno: gno.land/p/filetest/issue28251.T{}.m6 undefined (type gno.land/p/filetest/issue28251.T has no field or method m6)
// line 63: gc expects: invalid receiver type (\*\*T|T9); gno: invalid receiver type **gno.land/p/filetest/issue28251.T (base type is pointer type)

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 44; over-strict: lines 26, 27, 35, 36, 55; internal errors: lines 30, 48
