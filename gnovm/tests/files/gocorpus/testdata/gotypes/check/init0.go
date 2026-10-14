// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// initialization cycles

package init0

// initialization cycles (we don't know the types)
const (
	s0 /* ERROR "initialization cycle: s0 refers to itself" */ = s0

	x0 /* ERROR "initialization cycle for x0" */ = y0
	y0 = x0

	a0 = b0
	b0 /* ERROR "initialization cycle for b0" */ = c0
	c0 = d0
	d0 = b0
)

var (
	s1 /* ERROR "initialization cycle: s1 refers to itself" */ = s1

	x1 /* ERROR "initialization cycle for x1" */ = y1
	y1 = x1

	a1 = b1
	b1 /* ERROR "initialization cycle for b1" */ = c1
	c1 = d1
	d1 = b1
)

// initialization cycles (we know the types)
const (
	s2 /* ERROR "initialization cycle: s2 refers to itself" */ int = s2

	x2 /* ERROR "initialization cycle for x2" */ int = y2
	y2 = x2

	a2 = b2
	b2 /* ERROR "initialization cycle for b2" */ int = c2
	c2 = d2
	d2 = b2
)

var (
	s3 /* ERROR "initialization cycle: s3 refers to itself" */ int = s3

	x3 /* ERROR "initialization cycle for x3" */ int = y3
	y3 = x3

	a3 = b3
	b3 /* ERROR "initialization cycle for b3" */ int = c3
	c3 = d3
	d3 = b3
)

// cycles via struct fields

type S1 struct {
	f int
}
const cx3 S1 /* ERROR "invalid constant type" */ = S1{cx3.f}
var vx3 /* ERROR "initialization cycle: vx3 refers to itself" */ S1 = S1{vx3.f}

// cycles via functions

var x4 = x5
var x5 /* ERROR "initialization cycle for x5" */ = f1()
func f1() int { return x5*10 }

var x6, x7 /* ERROR "initialization cycle" */ = f2()
var x8 = x7
func f2() (int, int) { return f3() + f3(), 0 }
func f3() int { return x8 }

// cycles via function literals

var x9 /* ERROR "initialization cycle: x9 refers to itself" */ = func() int { return x9 }()

var x10 /* ERROR "initialization cycle for x10" */ = f4()

func f4() int {
	_ = func() {
		_ = x10
	}
	return 0
}

// cycles via method expressions

type T1 struct{}

func (T1) m() bool { _ = x11; return false }

var x11 /* ERROR "initialization cycle for x11" */ = T1.m(T1{})

// cycles via method values

type T2 struct{}

func (T2) m() bool { _ = x12; return false }

var t1 T2
var x12 /* ERROR "initialization cycle for x12" */ = t1.m

// GnoError:
// line 11: invalid recursive value: s0 -> s0
// line 13: invalid recursive value: x0 -> y0 -> x0
// line 17: invalid recursive value: b0 -> c0 -> d0 -> b0
// line 23: invalid recursive value: s1 -> s1
// line 25: invalid recursive value: x1 -> y1 -> x1
// line 29: invalid recursive value: b1 -> c1 -> d1 -> b1
// line 36: invalid recursive value: s2 -> s2
// line 38: invalid recursive value: x2 -> y2 -> x2
// line 42: invalid recursive value: b2 -> c2 -> d2 -> b2
// line 48: invalid recursive value: s3 -> s3
// line 50: invalid recursive value: x3 -> y3 -> x3
// line 54: invalid recursive value: b3 -> c3 -> d3 -> b3
// line 64: invalid recursive value: cx3 -> cx3
// line 65: invalid recursive value: vx3 -> vx3
// line 70: initialization cycle: x5 refers to f1 refers to x5
// line 73: initialization cycle: x6 refers to f2 refers to f3 refers to x8 refers to x7
// line 80: initialization cycle: x9 refers to x9
// line 82: initialization cycle: x10 refers to f4 refers to x10
// line 97: initialization cycle: x11 refers to T1.m refers to x11
// line 106: initialization cycle: x12 refers to T2.m refers to x12

// GoTypeCheckError:
// line 11
// line 13
// line 14
// line 17
// line 18
// line 19
// line 23
// line 25
// line 26
// line 29
// line 30
// line 31
// line 36
// line 38
// line 39
// line 42
// line 43
// line 44
// line 48
// line 50
// line 51
// line 54
// line 55
// line 56
// line 64
// line 65
// line 70
// line 71
// line 73
// line 74
// line 75
// line 76
// line 80
// line 82
// line 84
// line 95
// line 97
// line 103
// line 106

// GnoWordingDiffers:
// line 11: gc expects: initialization cycle: s0 refers to itself; gno: invalid recursive value: s0 -> s0
// line 13: gc expects: initialization cycle for x0; gno: invalid recursive value: x0 -> y0 -> x0
// line 17: gc expects: initialization cycle for b0; gno: invalid recursive value: b0 -> c0 -> d0 -> b0
// line 23: gc expects: initialization cycle: s1 refers to itself; gno: invalid recursive value: s1 -> s1
// line 25: gc expects: initialization cycle for x1; gno: invalid recursive value: x1 -> y1 -> x1
// line 29: gc expects: initialization cycle for b1; gno: invalid recursive value: b1 -> c1 -> d1 -> b1
// line 36: gc expects: initialization cycle: s2 refers to itself; gno: invalid recursive value: s2 -> s2
// line 38: gc expects: initialization cycle for x2; gno: invalid recursive value: x2 -> y2 -> x2
// line 42: gc expects: initialization cycle for b2; gno: invalid recursive value: b2 -> c2 -> d2 -> b2
// line 48: gc expects: initialization cycle: s3 refers to itself; gno: invalid recursive value: s3 -> s3
// line 50: gc expects: initialization cycle for x3; gno: invalid recursive value: x3 -> y3 -> x3
// line 54: gc expects: initialization cycle for b3; gno: invalid recursive value: b3 -> c3 -> d3 -> b3
// line 64: gc expects: invalid constant type; gno: invalid recursive value: cx3 -> cx3
// line 65: gc expects: initialization cycle: vx3 refers to itself; gno: invalid recursive value: vx3 -> vx3
// line 70: gc expects: initialization cycle for x5; gno: initialization cycle: x5 refers to f1 refers to x5
// line 80: gc expects: initialization cycle: x9 refers to itself; gno: initialization cycle: x9 refers to x9
// line 82: gc expects: initialization cycle for x10; gno: initialization cycle: x10 refers to f4 refers to x10
// line 97: gc expects: initialization cycle for x11; gno: initialization cycle: x11 refers to T1.m refers to x11
// line 106: gc expects: initialization cycle for x12; gno: initialization cycle: x12 refers to T2.m refers to x12

// Verdict: Clean errorcheck
