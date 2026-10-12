// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _(a, b, c int) {
	_ = a
	_ = a, b    // ERROR "assignment mismatch: 1 variable but 2 values"
	_ = a, b, c // ERROR "assignment mismatch: 1 variable but 3 values"

	_, _ = a // ERROR "assignment mismatch: 2 variables but 1 value"
	_, _ = a, b
	_, _ = a, b, c // ERROR "assignment mismatch: 2 variables but 3 values"

	_, _, _ = a    // ERROR "assignment mismatch: 3 variables but 1 value"
	_, _, _ = a, b // ERROR "assignment mismatch: 3 variables but 2 values"
	_, _, _ = a, b, c
}

func f1() int
func f2() (int, int)
func f3() (int, int, int)

func _() {
	_ = f1()
	_ = f2() // ERROR "assignment mismatch: 1 variable but f2 returns 2 values"
	_ = f3() // ERROR "assignment mismatch: 1 variable but f3 returns 3 values"

	_, _ = f1() // ERROR "assignment mismatch: 2 variables but f1 returns 1 value"
	_, _ = f2()
	_, _ = f3() // ERROR "assignment mismatch: 2 variables but f3 returns 3 values"

	_, _, _ = f1() // ERROR "assignment mismatch: 3 variables but f1 returns 1 value"
	_, _, _ = f2() // ERROR "assignment mismatch: 3 variables but f2 returns 2 values"
	_, _, _ = f3()

	// test just a few := cases as they use the same code as the = case
	a1 := f3()         // ERROR "assignment mismatch: 1 variable but f3 returns 3 values"
	a2, b2 := f1()     // ERROR "assignment mismatch: 2 variables but f1 returns 1 value"
	a3, b3, c3 := f2() // ERROR "assignment mismatch: 3 variables but f2 returns 2 values"

	_ = a1
	_, _ = a2, b2
	_, _, _ = a3, b3, c3
}

type T struct{}

func (T) f1() int
func (T) f2() (int, int)
func (T) f3() (int, int, int)

func _(x T) {
	_ = x.f1()
	_ = x.f2() // ERROR "assignment mismatch: 1 variable but .\.f2 returns 2 values"
	_ = x.f3() // ERROR "assignment mismatch: 1 variable but .\.f3 returns 3 values"

	_, _ = x.f1() // ERROR "assignment mismatch: 2 variables but .\.f1 returns 1 value"
	_, _ = x.f2()
	_, _ = x.f3() // ERROR "assignment mismatch: 2 variables but .\.f3 returns 3 values"

	_, _, _ = x.f1() // ERROR "assignment mismatch: 3 variables but .\.f1 returns 1 value"
	_, _, _ = x.f2() // ERROR "assignment mismatch: 3 variables but .\.f2 returns 2 values"
	_, _, _ = x.f3()

	// test just a few := cases as they use the same code as the = case
	a1 := x.f3()         // ERROR "assignment mismatch: 1 variable but .\.f3 returns 3 values"
	a2, b2 := x.f1()     // ERROR "assignment mismatch: 2 variables but .\.f1 returns 1 value"
	a3, b3, c3 := x.f2() // ERROR "assignment mismatch: 3 variables but .\.f2 returns 2 values"

	_ = a1
	_, _ = a2, b2
	_, _, _ = a3, b3, c3
}

// some one-off cases
func _() {
	_ = (f2)
	_ = f1(), 2         // ERROR "assignment mismatch: 1 variable but 2 values"
	_, _ = (f1()), f2() // ERROR "multiple-value f2\(\) .*in single-value context"
	_, _, _ = f3(), 3   // ERROR "assignment mismatch: 3 variables but 2 values|multiple-value f3\(\) .*in single-value context"
}

// GnoError:
// line 11: Machine.EvalStaticTypeOf(x) expression not yet preprocessed: b<VPInvalid(0)>
// line 12: Machine.EvalStaticTypeOf(x) expression not yet preprocessed: b<VPInvalid(0)>
// line 14: RHS should not be a<VPBlock(1,0)> when len(Lhs) > len(Rhs)
// line 16: Machine.EvalStaticTypeOf(x) expression not yet preprocessed: c<VPInvalid(0)>
// line 18: RHS should not be a<VPBlock(1,0)> when len(Lhs) > len(Rhs)
// line 19: assignment mismatch: 3 variables but 2 values
// line 23: function f1 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 24: function f2 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 25: function f3 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 45: name a1 not declared
// line 46: name a2 not declared
// line 47: name a3 not declared
// line 58: typleType has no property called named
// line 59: typleType has no property called named
// line 61: assignment mismatch: 2 variables but x<VPBlock(1,0)>.f1 returns 1 values
// line 63: assignment mismatch: 2 variables but x<VPBlock(1,0)>.f3 returns 3 values
// line 65: assignment mismatch: 3 variables but x<VPBlock(1,0)>.f1 returns 1 values
// line 66: assignment mismatch: 3 variables but x<VPBlock(1,0)>.f2 returns 2 values
// line 70: multiple-value x<VPBlock(1,0)>.f3 (value of type [int int int]) in single-value context
// line 71: assignment mismatch: 2 variables but x<VPBlock(1,0)>.f1 returns 1 values
// line 72: assignment mismatch: 3 variables but x<VPBlock(1,0)>.f2 returns 2 values

// GoTypeCheckError:
// line 11
// line 12
// line 14
// line 16
// line 18
// line 19
// line 29
// line 30
// line 32
// line 34
// line 36
// line 37
// line 41
// line 42
// line 43
// line 58
// line 59
// line 61
// line 63
// line 65
// line 66
// line 70
// line 71
// line 72
// line 82
// line 83
// line 84

// GoTypesOnly:
// line 29: gc expects: assignment mismatch: 1 variable but f2 returns 2 values
// line 30: gc expects: assignment mismatch: 1 variable but f3 returns 3 values
// line 32: gc expects: assignment mismatch: 2 variables but f1 returns 1 value
// line 34: gc expects: assignment mismatch: 2 variables but f3 returns 3 values
// line 36: gc expects: assignment mismatch: 3 variables but f1 returns 1 value
// line 37: gc expects: assignment mismatch: 3 variables but f2 returns 2 values
// line 41: gc expects: assignment mismatch: 1 variable but f3 returns 3 values
// line 42: gc expects: assignment mismatch: 2 variables but f1 returns 1 value
// line 43: gc expects: assignment mismatch: 3 variables but f2 returns 2 values
// line 82: gc expects: assignment mismatch: 1 variable but 2 values
// line 83: gc expects an error (no pattern)
// line 84: gc expects an error (no pattern)

// GnoOverStrictError:
// line 23: function f1 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 24: function f2 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 25: function f3 does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 45: name a1 not declared
// line 46: name a2 not declared
// line 47: name a3 not declared

// Verdict: KnownIssue errorcheck: 6 over-strict line(s), 0 uncaught marker(s)
