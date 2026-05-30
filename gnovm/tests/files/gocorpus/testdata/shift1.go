// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test illegal shifts.
// Issue 1708, illegal cases.
// Does not compile.

package p

func f(x int) int         { return 0 }
func g(x interface{}) int { return 0 }
func h(x float64) int     { return 0 }

// from the spec
var (
	s uint    = 33
	u         = 1.0 << s // ERROR "invalid operation|shift of non-integer operand"
	v float32 = 1 << s   // ERROR "invalid"
)

// non-constant shift expressions
var (
	e1       = g(2.0 << s) // ERROR "invalid|shift of non-integer operand"
	f1       = h(2 << s)   // ERROR "invalid"
	g1 int64 = 1.1 << s    // ERROR "truncated|must be integer"
)

// constant shift expressions
const c uint = 65

var (
	a2 int = 1.0 << c    // ERROR "overflow"
	b2     = 1.0 << c    // ERROR "overflow"
	d2     = f(1.0 << c) // ERROR "overflow"
)

var (
	// issues 4882, 4936.
	a3 = 1.0<<s + 0 // ERROR "invalid|shift of non-integer operand"
	// issue 4937
	b3 = 1<<s + 1 + 1.0 // ERROR "invalid|shift of non-integer operand"
	// issue 5014
	c3     = complex(1<<s, 0) // ERROR "invalid|shift of type float64"
	d3 int = complex(1<<s, 3) // ERROR "non-integer|cannot use.*as type int" "shift of type float64|must be integer"
	e3     = real(1 << s)     // ERROR "invalid"
	f3     = imag(1 << s)     // ERROR "invalid"
)

// from the spec
func _() {
	var (
		s uint  = 33
		i       = 1 << s         // 1 has type int
		j int32 = 1 << s         // 1 has type int32; j == 0
		k       = uint64(1 << s) // 1 has type uint64; k == 1<<33
		m int   = 1.0 << s       // 1.0 has type int
		n       = 1.0<<s != i    // 1.0 has type int; n == false if ints are 32bits in size
		o       = 1<<s == 2<<s   // 1 and 2 have type int; o == true if ints are 32bits in size
		// next test only fails on 32bit systems
		// p = 1<<s == 1<<33  // illegal if ints are 32bits in size: 1 has type int, but 1<<33 overflows int
		u          = 1.0 << s    // ERROR "non-integer|float64"
		u1         = 1.0<<s != 0 // ERROR "non-integer|float64"
		u2         = 1<<s != 1.0 // ERROR "non-integer|float64"
		v  float32 = 1 << s      // ERROR "non-integer|float32"
		w  int64   = 1.0 << 33   // 1.0<<33 is a constant shift expression

		_, _, _, _, _, _, _, _, _, _ = j, k, m, n, o, u, u1, u2, v, w
	)

	// non constants arguments trigger a different path
	f2 := 1.2
	s2 := "hi"
	_ = f2 << 2 // ERROR "shift of type float64|non-integer|must be integer"
	_ = s2 << 2 // ERROR "shift of type string|non-integer|must be integer"
}

// shifts in comparisons w/ untyped operands
var (
	_ = 1<<s == 1
	_ = 1<<s == 1.  // ERROR "invalid|shift of type float64"
	_ = 1.<<s == 1  // ERROR "invalid|shift of type float64"
	_ = 1.<<s == 1. // ERROR "invalid|non-integer|shift of type float64"

	_ = 1<<s+1 == 1
	_ = 1<<s+1 == 1.   // ERROR "invalid|shift of type float64"
	_ = 1<<s+1. == 1   // ERROR "invalid|shift of type float64"
	_ = 1<<s+1. == 1.  // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1 == 1   // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1 == 1.  // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1. == 1  // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1. == 1. // ERROR "invalid|non-integer|shift of type float64"

	_ = 1<<s == 1<<s
	_ = 1<<s == 1.<<s  // ERROR "invalid|shift of type float64"
	_ = 1.<<s == 1<<s  // ERROR "invalid|shift of type float64"
	_ = 1.<<s == 1.<<s // ERROR "invalid|non-integer|shift of type float64"

	_ = 1<<s+1<<s == 1
	_ = 1<<s+1<<s == 1.   // ERROR "invalid|shift of type float64"
	_ = 1<<s+1.<<s == 1   // ERROR "invalid|shift of type float64"
	_ = 1<<s+1.<<s == 1.  // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1<<s == 1   // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1<<s == 1.  // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1.<<s == 1  // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1.<<s == 1. // ERROR "invalid|non-integer|shift of type float64"

	_ = 1<<s+1<<s == 1<<s+1<<s
	_ = 1<<s+1<<s == 1<<s+1.<<s    // ERROR "invalid|shift of type float64"
	_ = 1<<s+1<<s == 1.<<s+1<<s    // ERROR "invalid|shift of type float64"
	_ = 1<<s+1<<s == 1.<<s+1.<<s   // ERROR "invalid|shift of type float64"
	_ = 1<<s+1.<<s == 1<<s+1<<s    // ERROR "invalid|shift of type float64"
	_ = 1<<s+1.<<s == 1<<s+1.<<s   // ERROR "invalid|shift of type float64"
	_ = 1<<s+1.<<s == 1.<<s+1<<s   // ERROR "invalid|shift of type float64"
	_ = 1<<s+1.<<s == 1.<<s+1.<<s  // ERROR "invalid|non-integer|shift of type float64"
	_ = 1.<<s+1<<s == 1<<s+1<<s    // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1<<s == 1<<s+1.<<s   // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1<<s == 1.<<s+1<<s   // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1<<s == 1.<<s+1.<<s  // ERROR "invalid|non-integer|shift of type float64"
	_ = 1.<<s+1.<<s == 1<<s+1<<s   // ERROR "invalid|shift of type float64"
	_ = 1.<<s+1.<<s == 1<<s+1.<<s  // ERROR "invalid|non-integer|shift of type float64"
	_ = 1.<<s+1.<<s == 1.<<s+1<<s  // ERROR "invalid|non-integer|shift of type float64"
	_ = 1.<<s+1.<<s == 1.<<s+1.<<s // ERROR "invalid|non-integer|shift of type float64"
)

// shifts in comparisons w/ typed operands
var (
	x int
	_ = 1<<s == x
	_ = 1.<<s == x
	_ = 1.1<<s == x // ERROR "truncated|must be integer"

	_ = 1<<s+x == 1
	_ = 1<<s+x == 1.
	_ = 1<<s+x == 1.1 // ERROR "truncated"
	_ = 1.<<s+x == 1
	_ = 1.<<s+x == 1.
	_ = 1.<<s+x == 1.1  // ERROR "truncated"
	_ = 1.1<<s+x == 1   // ERROR "truncated|must be integer"
	_ = 1.1<<s+x == 1.  // ERROR "truncated|must be integer"
	_ = 1.1<<s+x == 1.1 // ERROR "truncated|must be integer"

	_ = 1<<s == x<<s
	_ = 1.<<s == x<<s
	_ = 1.1<<s == x<<s // ERROR "truncated|must be integer"
)

// shifts as operands in non-arithmetic operations and as arguments
func _() {
	var s uint
	var a []int
	_ = a[1<<s]
	_ = a[1.]
	_ = a[1.<<s]
	_ = a[1.1<<s] // ERROR "integer|shift of type float64"

	_ = make([]int, 1)
	_ = make([]int, 1.)
	_ = make([]int, 1.<<s)
	_ = make([]int, 1.1<<s) // ERROR "non-integer|truncated|must be integer"

	_ = float32(1)
	_ = float32(1 << s) // ERROR "non-integer|shift of type float32|must be integer"
	_ = float32(1.)
	_ = float32(1. << s)  // ERROR "non-integer|shift of type float32|must be integer"
	_ = float32(1.1 << s) // ERROR "non-integer|shift of type float32|must be integer"

	_ = append(a, 1<<s)
	_ = append(a, 1.<<s)
	_ = append(a, 1.1<<s) // ERROR "truncated|must be integer"

	var b []float32
	_ = append(b, 1<<s)   // ERROR "non-integer|type float32"
	_ = append(b, 1.<<s)  // ERROR "non-integer|type float32"
	_ = append(b, 1.1<<s) // ERROR "non-integer|type float32|must be integer"

	_ = complex(1.<<s, 0)  // ERROR "non-integer|shift of type float64|must be integer"
	_ = complex(1.1<<s, 0) // ERROR "non-integer|shift of type float64|must be integer"
	_ = complex(0, 1.<<s)  // ERROR "non-integer|shift of type float64|must be integer"
	_ = complex(0, 1.1<<s) // ERROR "non-integer|shift of type float64|must be integer"

	var a4 float64
	var b4 int
	_ = complex(1<<s, a4) // ERROR "non-integer|shift of type float64|must be integer"
	_ = complex(1<<s, b4) // ERROR "invalid|non-integer|"

	var m1 map[int]string
	delete(m1, 1<<s)
	delete(m1, 1.<<s)
	delete(m1, 1.1<<s) // ERROR "truncated|shift of type float64|must be integer"

	var m2 map[float32]string
	delete(m2, 1<<s)   // ERROR "invalid|cannot use 1 << s as type float32"
	delete(m2, 1.<<s)  // ERROR "invalid|cannot use 1 << s as type float32"
	delete(m2, 1.1<<s) // ERROR "invalid|cannot use 1.1 << s as type float32"
}

// shifts of shifts
func _() {
	var s uint
	_ = 1 << (1 << s)
	_ = 1 << (1. << s)
	_ = 1 << (1.1 << s)   // ERROR "non-integer|truncated|must be integer"
	_ = 1. << (1 << s)    // ERROR "non-integer|shift of type float64|must be integer"
	_ = 1. << (1. << s)   // ERROR "non-integer|shift of type float64|must be integer"
	_ = 1.1 << (1.1 << s) // ERROR "invalid|non-integer|truncated"

	_ = (1 << s) << (1 << s)
	_ = (1 << s) << (1. << s)
	_ = (1 << s) << (1.1 << s)   // ERROR "truncated|must be integer"
	_ = (1. << s) << (1 << s)    // ERROR "non-integer|shift of type float64|must be integer"
	_ = (1. << s) << (1. << s)   // ERROR "non-integer|shift of type float64|must be integer"
	_ = (1.1 << s) << (1.1 << s) // ERROR "invalid|non-integer|truncated"

	var x int
	x = 1 << (1 << s)
	x = 1 << (1. << s)
	x = 1 << (1.1 << s) // ERROR "truncated|must be integer"
	x = 1. << (1 << s)
	x = 1. << (1. << s)
	x = 1.1 << (1.1 << s) // ERROR "truncated|must be integer"

	x = (1 << s) << (1 << s)
	x = (1 << s) << (1. << s)
	x = (1 << s) << (1.1 << s) // ERROR "truncated|must be integer"
	x = (1. << s) << (1 << s)
	x = (1. << s) << (1. << s)
	x = (1.1 << s) << (1.1 << s) // ERROR "truncated|must be integer"

	var y float32
	y = 1 << (1 << s)     // ERROR "non-integer|type float32"
	y = 1 << (1. << s)    // ERROR "non-integer|type float32"
	y = 1 << (1.1 << s)   // ERROR "invalid|truncated|float32"
	y = 1. << (1 << s)    // ERROR "non-integer|type float32"
	y = 1. << (1. << s)   // ERROR "non-integer|type float32"
	y = 1.1 << (1.1 << s) // ERROR "invalid|truncated|float32"

	var z complex128
	z = (1 << s) << (1 << s)     // ERROR "non-integer|type complex128"
	z = (1 << s) << (1. << s)    // ERROR "non-integer|type complex128"
	z = (1 << s) << (1.1 << s)   // ERROR "invalid|truncated|complex128"
	z = (1. << s) << (1 << s)    // ERROR "non-integer|type complex128|must be integer"
	z = (1. << s) << (1. << s)   // ERROR "non-integer|type complex128|must be integer"
	z = (1.1 << s) << (1.1 << s) // ERROR "invalid|truncated|complex128"

	_, _, _ = x, y, z
}

// GnoIncomplete: covered 62 of 105 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// GnoError:
// line 20: operator << not defined on: Float64Kind
// line 21: operator << not defined on: Float32Kind
// line 26: operator << not defined on: Float64Kind
// line 27: operator << not defined on: Float64Kind
// line 28: invalid operation: shifted operand (const (1.1 <untyped> bigdec)) (<untyped> bigdec) must be integer
// line 35: bigint overflows target kind
// line 36: bigint overflows target kind
// line 37: bigint overflows target kind
// line 42: operator << not defined on: Float64Kind
// line 44: operator << not defined on: BigdecKind
// line 46: name complex not defined in fileset with files [shift1.go]
// line 47: name complex not defined in fileset with files [shift1.go]
// line 48: name real not defined in fileset with files [shift1.go]
// line 49: name imag not defined in fileset with files [shift1.go]
// line 64: operator << not defined on: Float64Kind
// line 65: operator << not defined on: Float64Kind
// line 66: operator << not defined on: Float64Kind
// line 67: operator << not defined on: Float32Kind
// line 83: operator << not defined on: Float64Kind
// line 84: operator << not defined on: Float64Kind
// line 85: operator << not defined on: Float64Kind
// line 88: operator << not defined on: Float64Kind
// line 90: operator << not defined on: Float64Kind
// line 91: operator << not defined on: Float64Kind
// line 92: operator << not defined on: Float64Kind
// line 93: operator << not defined on: Float64Kind
// line 94: operator << not defined on: Float64Kind
// line 97: incompatible types in binary expression: <untyped> bigint EQL <untyped> bigdec
// line 98: incompatible types in binary expression: <untyped> bigdec EQL <untyped> bigint
// line 99: operator << not defined on: Float64Kind
// line 103: incompatible types in binary expression: <untyped> bigint ADD <untyped> bigdec
// line 104: incompatible types in binary expression: <untyped> bigint ADD <untyped> bigdec
// line 105: incompatible types in binary expression: <untyped> bigdec ADD <untyped> bigint
// line 106: incompatible types in binary expression: <untyped> bigdec ADD <untyped> bigint
// line 107: operator << not defined on: Float64Kind
// line 108: operator << not defined on: Float64Kind
// line 111: incompatible types in binary expression: <untyped> bigint ADD <untyped> bigdec
// line 112: incompatible types in binary expression: <untyped> bigdec ADD <untyped> bigint
// line 113: operator << not defined on: Float64Kind
// line 114: incompatible types in binary expression: <untyped> bigint ADD <untyped> bigdec
// line 115: incompatible types in binary expression: <untyped> bigint ADD <untyped> bigdec
// line 116: incompatible types in binary expression: <untyped> bigint ADD <untyped> bigdec
// line 117: incompatible types in binary expression: <untyped> bigint ADD <untyped> bigdec
// line 118: incompatible types in binary expression: <untyped> bigdec ADD <untyped> bigint
// line 119: incompatible types in binary expression: <untyped> bigdec ADD <untyped> bigint
// line 120: incompatible types in binary expression: <untyped> bigdec ADD <untyped> bigint
// line 121: incompatible types in binary expression: <untyped> bigdec ADD <untyped> bigint
// line 122: operator << not defined on: Float64Kind
// line 123: operator << not defined on: Float64Kind
// line 124: operator << not defined on: Float64Kind
// line 125: operator << not defined on: Float64Kind
// line 133: invalid operation: shifted operand (const (1.1 <untyped> bigdec)) (<untyped> bigdec) must be integer
// line 137: cannot convert untyped bigdec to integer -- 1.1 not an exact integer
// line 140: cannot convert untyped bigdec to integer -- 1.1 not an exact integer
// line 141: invalid operation: shifted operand (const (1.1 <untyped> bigdec)) (<untyped> bigdec) must be integer
// line 142: invalid operation: shifted operand (const (1.1 <untyped> bigdec)) (<untyped> bigdec) must be integer
// line 143: invalid operation: shifted operand (const (1.1 <untyped> bigdec)) (<untyped> bigdec) must be integer
// line 147: invalid operation: shifted operand (const (1.1 <untyped> bigdec)) (<untyped> bigdec) must be integer

// GoTypeCheckError:
// line 20: invalid operation: shifted operand 1.0 (type float64) must be integer
// line 21: invalid operation: shifted operand 1 (type float32) must be integer
// line 26: invalid operation: shifted operand 2.0 (type float64) must be integer
// line 27: invalid operation: shifted operand 2 (type float64) must be integer
// line 28: invalid operation: shifted operand 1.1 (untyped float constant) must be integer
// line 35: cannot use 1.0 << c (untyped int constant 36893488147419103232) as int value in variable declaration (overflows)
// line 36: cannot use 1.0 << c (untyped int constant 36893488147419103232) as int value in variable declaration (overflows)
// line 37: cannot use 1.0 << c (untyped int constant 36893488147419103232) as int value in argument to f (overflows)
// line 42: invalid operation: shifted operand 1.0 (type float64) must be integer
// line 44: invalid operation: shifted operand 1 (type float64) must be integer
// line 46: invalid operation: shifted operand 1 (type float64) must be integer
// line 47: invalid operation: shifted operand 1 (type float64) must be integer
// line 48: invalid operation: shifted operand 1 (type complex128) must be integer
// line 49: invalid operation: shifted operand 1 (type complex128) must be integer
// line 64: invalid operation: shifted operand 1.0 (type float64) must be integer
// line 65: invalid operation: shifted operand 1.0 (type float64) must be integer
// line 66: invalid operation: shifted operand 1 (type float64) must be integer
// line 67: invalid operation: shifted operand 1 (type float32) must be integer
// line 76: invalid operation: shifted operand f2 (variable of type float64) must be integer
// line 77: invalid operation: shifted operand s2 (variable of type string) must be integer
// line 83: invalid operation: shifted operand 1 (type float64) must be integer
// line 84: invalid operation: shifted operand 1. (type float64) must be integer
// line 85: invalid operation: shifted operand 1. (type float64) must be integer
// line 88: invalid operation: shifted operand 1 (type float64) must be integer
// line 89: invalid operation: shifted operand 1 (type float64) must be integer
// line 90: invalid operation: shifted operand 1 (type float64) must be integer
// line 91: invalid operation: shifted operand 1. (type float64) must be integer
// line 92: invalid operation: shifted operand 1. (type float64) must be integer
// line 93: invalid operation: shifted operand 1. (type float64) must be integer
// line 94: invalid operation: shifted operand 1. (type float64) must be integer
// line 97: invalid operation: shifted operand 1 (type float64) must be integer
// line 98: invalid operation: shifted operand 1. (type float64) must be integer
// line 99: invalid operation: shifted operand 1. (type float64) must be integer
// line 102: invalid operation: shifted operand 1 (type float64) must be integer
// line 103: invalid operation: shifted operand 1 (type float64) must be integer
// line 104: invalid operation: shifted operand 1 (type float64) must be integer
// line 105: invalid operation: shifted operand 1. (type float64) must be integer
// line 106: invalid operation: shifted operand 1. (type float64) must be integer
// line 107: invalid operation: shifted operand 1. (type float64) must be integer
// line 108: invalid operation: shifted operand 1. (type float64) must be integer
// line 111: invalid operation: shifted operand 1 (type float64) must be integer
// line 112: invalid operation: shifted operand 1 (type float64) must be integer
// line 113: invalid operation: shifted operand 1 (type float64) must be integer
// line 114: invalid operation: shifted operand 1 (type float64) must be integer
// line 115: invalid operation: shifted operand 1 (type float64) must be integer
// line 116: invalid operation: shifted operand 1 (type float64) must be integer
// line 117: invalid operation: shifted operand 1 (type float64) must be integer
// line 118: invalid operation: shifted operand 1. (type float64) must be integer
// line 119: invalid operation: shifted operand 1. (type float64) must be integer
// line 120: invalid operation: shifted operand 1. (type float64) must be integer
// line 121: invalid operation: shifted operand 1. (type float64) must be integer
// line 122: invalid operation: shifted operand 1. (type float64) must be integer
// line 123: invalid operation: shifted operand 1. (type float64) must be integer
// line 124: invalid operation: shifted operand 1. (type float64) must be integer
// line 125: invalid operation: shifted operand 1. (type float64) must be integer
// line 133: invalid operation: shifted operand 1.1 (untyped float constant) must be integer
// line 137: 1.1 (untyped float constant) truncated to int
// line 140: 1.1 (untyped float constant) truncated to int
// line 141: invalid operation: shifted operand 1.1 (untyped float constant) must be integer
// line 142: invalid operation: shifted operand 1.1 (untyped float constant) must be integer
// line 143: invalid operation: shifted operand 1.1 (untyped float constant) must be integer
// line 147: invalid operation: shifted operand 1.1 (untyped float constant) must be integer
