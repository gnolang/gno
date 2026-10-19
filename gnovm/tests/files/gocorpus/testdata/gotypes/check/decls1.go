// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// variable declarations

package decls1

import (
	"math"
)

// Global variables without initialization
var (
	a, b bool
	c byte
	d uint8
	r rune
	i int
	j, k, l int
	x, y float32
	xx, yy float64
	u, v complex64
	uu, vv complex128
	s, t string
	array []byte
	iface interface{}

	blank _ /* ERROR "cannot use _" */
)

// Global variables with initialization
var (
	s1 = i + j
	s2 = i /* ERROR "mismatched types" */ + x
	s3 = c + d
	s4 = s + t
	s5 = s /* ERROR "invalid operation" */ / t
	s6 = array[t1]
	s7 = array[x /* ERROR "integer" */]
	s8 = &a
	s10 = &42 /* ERROR "cannot take address" */
	s11 = &v
	s12 = -(u + *t11) / *&v
	s13 = a /* ERROR "shifted operand" */ << d
	s14 = i << j
	s18 = math.Pi * 10.0
	s19 = s1 /* ERROR "cannot call" */ ()
 	s20 = f0 /* ERROR "no value" */ ()
	s21 = f6(1, s1, i)
	s22 = f6(1, s1, uu /* ERRORx `cannot use .* in argument` */ )

	t1 int = i + j
	t2 int = i /* ERROR "mismatched types" */ + x
	t3 int = c /* ERRORx `cannot use .* variable declaration` */ + d
	t4 string = s + t
	t5 string = s /* ERROR "invalid operation" */ / t
	t6 byte = array[t1]
	t7 byte = array[x /* ERROR "must be integer" */]
	t8 *int = & /* ERRORx `cannot use .* variable declaration` */ a
	t10 *int = &42 /* ERROR "cannot take address" */
	t11 *complex64 = &v
	t12 complex64 = -(u + *t11) / *&v
	t13 int = a /* ERROR "shifted operand" */ << d
	t14 int = i << j
	t15 math /* ERROR "math is not a type" */
	t16 math.xxx /* ERROR "undefined" */
	t17 math /* ERROR "not a type" */ .Pi
	t18 float64 = math.Pi * 10.0
	t19 int = t1 /* ERROR "cannot call" */ ()
	t20 int = f0 /* ERROR "no value" */ ()
	t21 int = a /* ERRORx `cannot use .* variable declaration` */
)

// Various more complex expressions
var (
	u1 = x /* ERROR "not an interface" */ .(int)
	u2 = iface.([]int)
	u3 = iface.(a /* ERROR "not a type" */ )
	u4, ok = iface.(int)
	u5, ok2, ok3 = iface /* ERROR "assignment mismatch" */ .(int)
)

// Constant expression initializations
var (
	v1 = 1 /* ERROR "mismatched types untyped int and untyped string" */ + "foo"
	v2 = c + 255
	v3 = c + 256 /* ERROR "overflows" */
	v4 = r + 2147483647
	v5 = r + 2147483648 /* ERROR "overflows" */
	v6 = 42
	v7 = v6 + 9223372036854775807
	v8 = v6 + 9223372036854775808 /* ERROR "overflows" */
	v9 = i + 1 << 10
	v10 byte = 1024 /* ERROR "overflows" */
	v11 = xx/yy*yy - xx
	v12 = true && false
	v13 = nil /* ERROR "use of untyped nil" */
	v14 string = 257 // ERRORx `cannot use 257 .* as string value in variable declaration$`
	v15 int8 = 257 // ERRORx `cannot use 257 .* as int8 value in variable declaration .*overflows`
)

// Multiple assignment expressions
var (
	m1a, m1b = 1, 2
	m2a, m2b, m2c /* ERROR "missing init expr for m2c" */ = 1, 2
	m3a, m3b = 1, 2, 3 /* ERROR "extra init expr 3" */
)

func _() {
	var (
		m1a, m1b = 1, 2
		m2a, m2b, m2c /* ERROR "missing init expr for m2c" */ = 1, 2
		m3a, m3b = 1, 2, 3 /* ERROR "extra init expr 3" */
	)

	_, _ = m1a, m1b
	_, _, _ = m2a, m2b, m2c
	_, _ = m3a, m3b
}

// Declaration of parameters and results
func f0() {}
func f1(a /* ERROR "not a type" */) {}
func f2(a, b, c d /* ERROR "not a type" */) {}

func f3() int { return 0 }
func f4() a /* ERROR "not a type" */ { return 0 }
func f5() (a, b, c d /* ERROR "not a type" */) { return }

func f6(a, b, c int) complex128 { return 0 }

// Declaration of receivers
type T struct{}

func (T) m0() {}
func (*T) m1() {}
func (x T) m2() {}
func (x *T) m3() {}

// Initialization functions
func init() {}
func init /* ERROR "no arguments and no return values" */ (int) {}
func init /* ERROR "no arguments and no return values" */ () int { return 0 }
func init /* ERROR "no arguments and no return values" */ (int) int { return 0 }
func (T) init(int) int { return 0 }

// GnoError:
// line 23: name not defined: complex64
// line 24: name not defined: complex128
// line 29: cannot use _ as value or type
// line 35: invalid operation: i<VPBlock(2,5)> + x<VPBlock(2,9)> (mismatched types int and float32)
// line 38: operator / not defined on: StringKind
// line 40: expected typed integer kind, but got Float32Kind
// line 45: operator << not defined on: BoolKind
// line 48: unexpected func type int (gnolang.PrimitiveType)
// line 49: f0<VPBlock(2,91)> (no value) used as value
// line 54: invalid operation: i<VPBlock(2,5)> + x<VPBlock(2,9)> (mismatched types int and float32)
// line 55: cannot use uint8 as int
// line 57: operator / not defined on: StringKind
// line 59: expected typed integer kind, but got Float32Kind
// line 60: cannot use bool as int
// line 62: name not defined: complex64
// line 63: name not defined: complex64
// line 64: operator << not defined on: BoolKind
// line 66: package math cannot only be referred to in a selector expression
// line 67: cannot access math.xxx from gno.land/p/filetest/decls1
// line 68: (const (3.1415926536 <untyped> bigdec)) is not a type
// line 70: unexpected func type int (gnolang.PrimitiveType)
// line 71: f0<VPBlock(2,91)> (no value) used as value
// line 72: cannot use bool as int
// line 77: invalid operation: x<VPBlock(2,9)> (variable of type float32) is not an interface
// line 79: a<VPBlock(2,0)> is not a type
// line 81: assignment mismatch: 3 variable(s) but iface<VPBlock(2,20)>.((const-type int)) returns 2 value(s)
// line 86: cannot use untyped Bigint as StringKind
// line 88: bigint overflows target kind
// line 90: bigint overflows target kind
// line 93: bigint overflows target kind
// line 95: bigint overflows target kind
// line 98: use of untyped nil in variable declaration
// line 99: cannot use untyped Bigint as StringKind
// line 100: bigint overflows target kind
// line 106: assignment mismatch: 3 variable(s) but 2 value(s)
// line 107: Machine.EvalStaticTypeOf(x) expression not yet preprocessed: 3
// line 113: missing init expr for m2c
// line 114: extra init expr 3
// line 124: a<VPBlock(2,0)> is not a type
// line 125: d<VPBlock(2,3)> is not a type
// line 128: a<VPBlock(2,0)> is not a type
// line 129: d<VPBlock(2,3)> is not a type
// line 131: name not defined: complex128
// line 143: func init must have no arguments and no return values
// line 144: func init must have no arguments and no return values
// line 145: func init must have no arguments and no return values

// GnoWordingDiffers:
// line 38: gc expects: invalid operation; gno: operator / not defined on: StringKind
// line 45: gc expects: shifted operand; gno: operator << not defined on: BoolKind
// line 48: gc expects: cannot call; gno: unexpected func type int (gnolang.PrimitiveType)
// line 55: gc expects: cannot use .* variable declaration; gno: cannot use uint8 as int
// line 57: gc expects: invalid operation; gno: operator / not defined on: StringKind
// line 59: gc expects: must be integer; gno: expected typed integer kind, but got Float32Kind
// line 60: gc expects: cannot use .* variable declaration; gno: cannot use bool as int
// line 64: gc expects: shifted operand; gno: operator << not defined on: BoolKind
// line 66: gc expects: math is not a type; gno: package math cannot only be referred to in a selector expression
// line 67: gc expects: undefined; gno: cannot access math.xxx from gno.land/p/filetest/decls1
// line 70: gc expects: cannot call; gno: unexpected func type int (gnolang.PrimitiveType)
// line 72: gc expects: cannot use .* variable declaration; gno: cannot use bool as int
// line 86: gc expects: mismatched types untyped int and untyped string; gno: cannot use untyped Bigint as StringKind
// line 99: gc expects: cannot use 257 .* as string value in variable declaration$; gno: cannot use untyped Bigint as StringKind
// line 100: gc expects: cannot use 257 .* as int8 value in variable declaration .*overflows; gno: bigint overflows target kind
// line 106: gc expects: missing init expr for m2c; gno: assignment mismatch: 3 variable(s) but 2 value(s)
// line 107: gc expects: extra init expr 3; gno: Machine.EvalStaticTypeOf(x) expression not yet preprocessed: 3

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 42, 51, 61; over-strict: lines 23, 24, 62, 63, 131
