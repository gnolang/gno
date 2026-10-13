// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// binary expressions

package expr1

type mybool bool

func _(x, y bool, z mybool) {
	x = x || y
	x = x || true
	x = x || false
	x = x && y
	x = x && true
	x = x && false

	z = z /* ERROR "mismatched types" */ || y
	z = z || true
	z = z || false
	z = z /* ERROR "mismatched types" */ && y
	z = z && true
	z = z && false
}

type myint int

func _(x, y int, z myint) {
	x = x + 1
	x = x + 1.0
	x = x + 1.1 // ERROR "truncated to int"
	x = x + y
	x = x - y
	x = x * y
	x = x / y
	x = x % y
	x = x << y
	x = x >> y

	z = z + 1
	z = z + 1.0
	z = z + 1.1 // ERROR "truncated to int"
	z = z /* ERROR "mismatched types" */ + y
	z = z /* ERROR "mismatched types" */ - y
	z = z /* ERROR "mismatched types" */ * y
	z = z /* ERROR "mismatched types" */ / y
	z = z /* ERROR "mismatched types" */ % y
	z = z << y
	z = z >> y
}

type myuint uint

func _(x, y uint, z myuint) {
	x = x + 1
	x = x + - /* ERROR "overflows uint" */ 1
	x = x + 1.0
	x = x + 1.1 // ERROR "truncated to uint"
	x = x + y
	x = x - y
	x = x * y
	x = x / y
	x = x % y
	x = x << y
	x = x >> y

	z = z + 1
	z = x + - /* ERROR "overflows uint" */ 1
	z = z + 1.0
	z = z + 1.1 // ERROR "truncated to uint"
	z = z /* ERROR "mismatched types" */ + y
	z = z /* ERROR "mismatched types" */ - y
	z = z /* ERROR "mismatched types" */ * y
	z = z /* ERROR "mismatched types" */ / y
	z = z /* ERROR "mismatched types" */ % y
	z = z << y
	z = z >> y
}

type myfloat64 float64

func _(x, y float64, z myfloat64) {
	x = x + 1
	x = x + -1
	x = x + 1.0
	x = x + 1.1
	x = x + y
	x = x - y
	x = x * y
	x = x / y
	x = x /* ERROR "not defined" */ % y
	x = x /* ERRORx `operand x .* must be integer` */ << y
	x = x /* ERRORx `operand x .* must be integer` */ >> y

	z = z + 1
	z = z + -1
	z = z + 1.0
	z = z + 1.1
	z = z /* ERROR "mismatched types" */ + y
	z = z /* ERROR "mismatched types" */ - y
	z = z /* ERROR "mismatched types" */ * y
	z = z /* ERROR "mismatched types" */ / y
	z = z /* ERROR "mismatched types" */ % y
	z = z /* ERRORx `operand z .* must be integer` */ << y
	z = z /* ERRORx `operand z .* must be integer` */ >> y
}

type mystring string

func _(x, y string, z mystring) {
	x = x + "foo"
	x = x /* ERROR "not defined" */ - "foo"
	x = x /* ERROR "mismatched types string and untyped int" */ + 1
	x = x + y
	x = x /* ERROR "not defined" */ - y
	x = x /* ERROR "mismatched types string and untyped int" */* 10
}

func f() (a, b int) { return }

func _(x int) {
	_ = f /* ERROR "multiple-value f" */ () + 1
	_ = x + f /* ERROR "multiple-value f" */ ()
	_ = f /* ERROR "multiple-value f" */ () + f
	_ = f /* ERROR "multiple-value f" */ () + f /* ERROR "multiple-value f" */ ()
}

// GnoError:
// line 19: invalid operation: z<VPBlock(1,2)> || y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.mybool and bool)
// line 22: invalid operation: z<VPBlock(1,2)> && y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.mybool and bool)
// line 32: cannot convert untyped bigdec to integer -- 1.1 not an exact integer
// line 43: cannot convert untyped bigdec to integer -- 1.1 not an exact integer
// line 44: invalid operation: z<VPBlock(1,2)> + y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myint and int)
// line 45: invalid operation: z<VPBlock(1,2)> - y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myint and int)
// line 46: invalid operation: z<VPBlock(1,2)> * y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myint and int)
// line 47: invalid operation: z<VPBlock(1,2)> / y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myint and int)
// line 48: invalid operation: z<VPBlock(1,2)> % y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myint and int)
// line 57: bigint underflows target kind
// line 59: cannot convert untyped bigdec to integer -- 1.1 not an exact integer
// line 69: bigint underflows target kind
// line 71: cannot convert untyped bigdec to integer -- 1.1 not an exact integer
// line 72: invalid operation: z<VPBlock(1,2)> + y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myuint and uint)
// line 73: invalid operation: z<VPBlock(1,2)> - y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myuint and uint)
// line 74: invalid operation: z<VPBlock(1,2)> * y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myuint and uint)
// line 75: invalid operation: z<VPBlock(1,2)> / y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myuint and uint)
// line 76: invalid operation: z<VPBlock(1,2)> % y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myuint and uint)
// line 92: operator % not defined on: Float64Kind
// line 93: invalid operation: invalid shift count: y<VPBlock(1,1)>
// line 94: invalid operation: invalid shift count: y<VPBlock(1,1)>
// line 100: invalid operation: z<VPBlock(1,2)> + y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myfloat64 and float64)
// line 101: invalid operation: z<VPBlock(1,2)> - y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myfloat64 and float64)
// line 102: invalid operation: z<VPBlock(1,2)> * y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myfloat64 and float64)
// line 103: invalid operation: z<VPBlock(1,2)> / y<VPBlock(1,1)> (mismatched types gno.land/p/filetest/expr1.myfloat64 and float64)
// line 104: operator % not defined on: Float64Kind
// line 105: invalid operation: invalid shift count: y<VPBlock(1,1)>
// line 106: invalid operation: invalid shift count: y<VPBlock(1,1)>
// line 113: operator - not defined on: StringKind
// line 114: cannot use untyped Bigint as StringKind
// line 116: operator - not defined on: StringKind
// line 117: operator * not defined on: StringKind
// line 123: operator + not defined on: TupleKind
// line 124: operator + not defined on: TupleKind
// line 125: operator + not defined on: FuncKind
// line 126: operator + not defined on: TupleKind

// GoTypeCheckError:
// line 19
// line 22
// line 32
// line 43
// line 44
// line 45
// line 46
// line 47
// line 48
// line 57
// line 59
// line 69
// line 71
// line 72
// line 73
// line 74
// line 75
// line 76
// line 92
// line 93
// line 94
// line 100
// line 101
// line 102
// line 103
// line 104
// line 105
// line 106
// line 113
// line 114
// line 116
// line 117
// line 123
// line 124
// line 125
// line 126

// Clean: errorcheck
