// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// initialization cycles

package init2

// cycles through functions

func f1() int { _ = x1; return 0 }
var x1 /* ERROR "initialization cycle" */ = f1

func f2() int { _ = x2; return 0 }
var x2 /* ERROR "initialization cycle" */ = f2()

// cycles through method expressions

type T3 int
func (T3) m() int { _ = x3; return 0 }
var x3 /* ERROR "initialization cycle" */ = T3.m

type T4 int
func (T4) m() int { _ = x4; return 0 }
var x4 /* ERROR "initialization cycle" */ = T4.m(0)

type T3p int
func (*T3p) m() int { _ = x3p; return 0 }
var x3p /* ERROR "initialization cycle" */ = (*T3p).m

type T4p int
func (*T4p) m() int { _ = x4p; return 0 }
var x4p /* ERROR "initialization cycle" */ = (*T4p).m(nil)

// cycles through method expressions of embedded methods

type T5 struct { E5 }
type E5 int
func (E5) m() int { _ = x5; return 0 }
var x5 /* ERROR "initialization cycle" */ = T5.m

type T6 struct { E6 }
type E6 int
func (E6) m() int { _ = x6; return 0 }
var x6 /* ERROR "initialization cycle" */ = T6.m(T6{0})

type T5p struct { E5p }
type E5p int
func (*E5p) m() int { _ = x5p; return 0 }
var x5p /* ERROR "initialization cycle" */ = (*T5p).m

type T6p struct { E6p }
type E6p int
func (*E6p) m() int { _ = x6p; return 0 }
var x6p /* ERROR "initialization cycle" */ = (*T6p).m(nil)

// cycles through method values

type T7 int
func (T7) m() int { _ = x7; return 0 }
var x7 /* ERROR "initialization cycle" */ = T7(0).m

type T8 int
func (T8) m() int { _ = x8; return 0 }
var x8 /* ERROR "initialization cycle" */ = T8(0).m()

type T7p int
func (*T7p) m() int { _ = x7p; return 0 }
var x7p /* ERROR "initialization cycle" */ = new(T7p).m

type T8p int
func (*T8p) m() int { _ = x8p; return 0 }
var x8p /* ERROR "initialization cycle" */ = new(T8p).m()

type T7v int
func (T7v) m() int { _ = x7v; return 0 }
var x7var T7v
var x7v /* ERROR "initialization cycle" */ = x7var.m

type T8v int
func (T8v) m() int { _ = x8v; return 0 }
var x8var T8v
var x8v /* ERROR "initialization cycle" */ = x8var.m()

type T7pv int
func (*T7pv) m() int { _ = x7pv; return 0 }
var x7pvar *T7pv
var x7pv /* ERROR "initialization cycle" */ = x7pvar.m

type T8pv int
func (*T8pv) m() int { _ = x8pv; return 0 }
var x8pvar *T8pv
var x8pv /* ERROR "initialization cycle" */ = x8pvar.m()

// cycles through method values of embedded methods

type T9 struct { E9 }
type E9 int
func (E9) m() int { _ = x9; return 0 }
var x9 /* ERROR "initialization cycle" */ = T9{0}.m

type T10 struct { E10 }
type E10 int
func (E10) m() int { _ = x10; return 0 }
var x10 /* ERROR "initialization cycle" */ = T10{0}.m()

type T9p struct { E9p }
type E9p int
func (*E9p) m() int { _ = x9p; return 0 }
var x9p /* ERROR "initialization cycle" */ = new(T9p).m

type T10p struct { E10p }
type E10p int
func (*E10p) m() int { _ = x10p; return 0 }
var x10p /* ERROR "initialization cycle" */ = new(T10p).m()

type T9v struct { E9v }
type E9v int
func (E9v) m() int { _ = x9v; return 0 }
var x9var T9v
var x9v /* ERROR "initialization cycle" */ = x9var.m

type T10v struct { E10v }
type E10v int
func (E10v) m() int { _ = x10v; return 0 }
var x10var T10v
var x10v /* ERROR "initialization cycle" */ = x10var.m()

type T9pv struct { E9pv }
type E9pv int
func (*E9pv) m() int { _ = x9pv; return 0 }
var x9pvar *T9pv
var x9pv /* ERROR "initialization cycle" */ = x9pvar.m

type T10pv struct { E10pv }
type E10pv int
func (*E10pv) m() int { _ = x10pv; return 0 }
var x10pvar *T10pv
var x10pv /* ERROR "initialization cycle" */ = x10pvar.m()

// GnoError:
// line 12: initialization cycle: x1 refers to f1 refers to x1
// line 15: initialization cycle: x2 refers to f2 refers to x2
// line 21: initialization cycle: x3 refers to T3.m refers to x3
// line 25: initialization cycle: x4 refers to T4.m refers to x4
// line 29: initialization cycle: x3p refers to T3p.m refers to x3p
// line 33: initialization cycle: x4p refers to T4p.m refers to x4p
// line 40: unknown *DeclaredType method named m
// line 45: unknown *DeclaredType method named m
// line 50: unknown *DeclaredType method named m
// line 55: unknown *DeclaredType method named m
// line 61: initialization cycle: x7 refers to T7.m refers to x7
// line 65: initialization cycle: x8 refers to T8.m refers to x8
// line 69: initialization cycle: x7p refers to T7p.m refers to x7p
// line 73: initialization cycle: x8p refers to T8p.m refers to x8p
// line 78: initialization cycle: x7v refers to T7v.m refers to x7v
// line 83: initialization cycle: x8v refers to T8v.m refers to x8v
// line 88: initialization cycle: x7pv refers to T7pv.m refers to x7pv
// line 93: initialization cycle: x8pv refers to T8pv.m refers to x8pv
// line 100: initialization cycle: x9 refers to E9.m refers to x9
// line 105: initialization cycle: x10 refers to E10.m refers to x10
// line 110: initialization cycle: x9p refers to E9p.m refers to x9p
// line 115: initialization cycle: x10p refers to E10p.m refers to x10p
// line 121: initialization cycle: x9v refers to E9v.m refers to x9v
// line 127: initialization cycle: x10v refers to E10v.m refers to x10v
// line 133: initialization cycle: x9pv refers to E9pv.m refers to x9pv
// line 139: initialization cycle: x10pv refers to E10pv.m refers to x10pv

// GoTypeCheckError:
// line 11
// line 12
// line 14
// line 15
// line 20
// line 21
// line 24
// line 25
// line 28
// line 29
// line 32
// line 33
// line 39
// line 40
// line 44
// line 45
// line 49
// line 50
// line 54
// line 55
// line 60
// line 61
// line 64
// line 65
// line 68
// line 69
// line 72
// line 73
// line 76
// line 78
// line 81
// line 83
// line 86
// line 88
// line 91
// line 93
// line 99
// line 100
// line 104
// line 105
// line 109
// line 110
// line 114
// line 115
// line 119
// line 121
// line 125
// line 127
// line 131
// line 133
// line 137
// line 139

// GnoWordingDiffers:
// line 40: gc expects: initialization cycle; gno: unknown *DeclaredType method named m
// line 45: gc expects: initialization cycle; gno: unknown *DeclaredType method named m
// line 50: gc expects: initialization cycle; gno: unknown *DeclaredType method named m
// line 55: gc expects: initialization cycle; gno: unknown *DeclaredType method named m

// Verdict: Clean errorcheck
