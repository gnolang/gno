// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package expr3

import "time"

func indexes() {
	_ = 1 /* ERROR "cannot index" */ [0]
	_ = indexes /* ERROR "cannot index" */ [0]
	_ = ( /* ERROR "cannot slice" */ 12 + 3)[1:2]

	var a [10]int
	_ = a[true /* ERROR "cannot convert" */ ]
	_ = a["foo" /* ERROR "cannot convert" */ ]
	_ = a[1.1 /* ERROR "truncated" */ ]
	_ = a[1.0]
	_ = a[- /* ERROR "negative" */ 1]
	_ = a[- /* ERROR "negative" */ 1 :]
	_ = a[: - /* ERROR "negative" */ 1]
	_ = a[: /* ERROR "middle index required" */ : /* ERROR "final index required" */ ]
	_ = a[0: /* ERROR "middle index required" */ : /* ERROR "final index required" */ ]
	_ = a[0: /* ERROR "middle index required" */ :10]
	_ = a[:10:10]

	var a0 int
	a0 = a[0]
	_ = a0
	var a1 int32
	a1 = a /* ERRORx `cannot use .* in assignment` */ [1]
	_ = a1

	_ = a[9]
	_ = a[10 /* ERRORx `index .* out of bounds` */ ]
	_ = a[1 /* ERROR "overflows" */ <<100]
	_ = a[1<< /* ERROR "constant shift overflow" */ 1000] // no out-of-bounds follow-on error
	_ = a[10:]
	_ = a[:10]
	_ = a[10:10]
	_ = a[11 /* ERRORx `index .* out of bounds` */ :]
	_ = a[: 11 /* ERRORx `index .* out of bounds` */ ]
	_ = a[: 1 /* ERROR "overflows" */ <<100]
	_ = a[:10:10]
	_ = a[:11 /* ERRORx `index .* out of bounds` */ :10]
	_ = a[:10:11 /* ERRORx `index .* out of bounds` */ ]
	_ = a[10:0 /* ERROR "invalid slice indices" */ :10]
	_ = a[0:10:0 /* ERROR "invalid slice indices" */ ]
	_ = a[10:0 /* ERROR "invalid slice indices" */:0]
	_ = &a /* ERROR "cannot take address" */ [:10]

	pa := &a
	_ = pa[9]
	_ = pa[10 /* ERRORx `index .* out of bounds` */ ]
	_ = pa[1 /* ERROR "overflows" */ <<100]
	_ = pa[10:]
	_ = pa[:10]
	_ = pa[10:10]
	_ = pa[11 /* ERRORx `index .* out of bounds` */ :]
	_ = pa[: 11 /* ERRORx `index .* out of bounds` */ ]
	_ = pa[: 1 /* ERROR "overflows" */ <<100]
	_ = pa[:10:10]
	_ = pa[:11 /* ERRORx `index .* out of bounds` */ :10]
	_ = pa[:10:11 /* ERRORx `index .* out of bounds` */ ]
	_ = pa[10:0 /* ERROR "invalid slice indices" */ :10]
	_ = pa[0:10:0 /* ERROR "invalid slice indices" */ ]
	_ = pa[10:0 /* ERROR "invalid slice indices" */ :0]
	_ = &pa /* ERROR "cannot take address" */ [:10]

	var b [0]int
	_ = b[0 /* ERRORx `index .* out of bounds` */ ]
	_ = b[:]
	_ = b[0:]
	_ = b[:0]
	_ = b[0:0]
	_ = b[0:0:0]
	_ = b[1 /* ERRORx `index .* out of bounds` */ :0:0]

	var s []int
	_ = s[- /* ERROR "negative" */ 1]
	_ = s[- /* ERROR "negative" */ 1 :]
	_ = s[: - /* ERROR "negative" */ 1]
	_ = s[0]
	_ = s[1:2]
	_ = s[2:1 /* ERROR "invalid slice indices" */ ]
	_ = s[2:]
	_ = s[: 1 /* ERROR "overflows" */ <<100]
	_ = s[1 /* ERROR "overflows" */ <<100 :]
	_ = s[1 /* ERROR "overflows" */ <<100 : 1 /* ERROR "overflows" */ <<100]
	_ = s[: /* ERROR "middle index required" */ :  /* ERROR "final index required" */ ]
	_ = s[:10:10]
	_ = s[10:0 /* ERROR "invalid slice indices" */ :10]
	_ = s[0:10:0 /* ERROR "invalid slice indices" */ ]
	_ = s[10:0 /* ERROR "invalid slice indices" */ :0]
	_ = &s /* ERROR "cannot take address" */ [:10]

	var m map[string]int
	_ = m[0 /* ERRORx `cannot use .* in map index` */ ]
	_ = m /* ERROR "cannot slice" */ ["foo" : "bar"]
	_ = m["foo"]
	// ok is of type bool
	type mybool bool
	var ok mybool
	_, ok = m["bar"]
	_ = ok
	_ = m/* ERROR "mismatched types int and untyped string" */[0 /* ERROR "cannot use 0" */ ] + "foo"

	var t string
	_ = t[- /* ERROR "negative" */ 1]
	_ = t[- /* ERROR "negative" */ 1 :]
	_ = t[: - /* ERROR "negative" */ 1]
	_ = t[1:2:3 /* ERROR "3-index slice of string" */ ]
	_ = "foo"[1:2:3 /* ERROR "3-index slice of string" */ ]
	var t0 byte
	t0 = t[0]
	_ = t0
	var t1 rune
	t1 = t /* ERRORx `cannot use .* in assignment` */ [2]
	_ = t1
	_ = ("foo" + "bar")[5]
	_ = ("foo" + "bar")[6 /* ERRORx `index .* out of bounds` */ ]

	const c = "foo"
	_ = c[- /* ERROR "negative" */ 1]
	_ = c[- /* ERROR "negative" */ 1 :]
	_ = c[: - /* ERROR "negative" */ 1]
	var c0 byte
	c0 = c[0]
	_ = c0
	var c2 float32
	c2 = c /* ERRORx `cannot use .* in assignment` */ [2]
	_ = c[3 /* ERRORx `index .* out of bounds` */ ]
	_ = ""[0 /* ERRORx `index .* out of bounds` */ ]
	_ = c2

	_ = s[1<<30] // no compile-time error here

	// issue 4913
	type mystring string
	var ss string
	var ms mystring
	var i, j int
	ss = "foo"[1:2]
	ss = "foo"[i:j]
	ms = "foo" /* ERRORx `cannot use .* in assignment` */ [1:2]
	ms = "foo" /* ERRORx `cannot use .* in assignment` */ [i:j]
	_, _ = ss, ms
}

type T struct {
	x int
	y func()
}

func (*T) m() {}

func method_expressions() {
	_ = T.a /* ERROR "no field or method" */
	_ = T.x /* ERROR "has no method" */
	_ = T.m /* ERROR "invalid method expression T.m (needs pointer receiver (*T).m)" */
	_ = (*T).m

	var f func(*T) = T.m /* ERROR "invalid method expression T.m (needs pointer receiver (*T).m)" */
	var g func(*T) = (*T).m
	_, _ = f, g

	_ = T.y /* ERROR "has no method" */
	_ = (*T).y /* ERROR "has no method" */
}

func struct_literals() {
	type T0 struct {
		a, b, c int
	}

	type T1 struct {
		T0
		a, b int
		u float64
		s string
	}

	// keyed elements
	_ = T1{}
	_ = T1{a: 0, 1 /* ERRORx `mixture of .* elements` */ }
	_ = T1{aa /* ERROR "unknown field" */ : 0}
	_ = T1{1 /* ERROR "invalid field name" */ : 0}
	_ = T1{a: 0, s: "foo", u: 0, a /* ERROR "duplicate field" */: 10}
	_ = T1{a: "foo" /* ERRORx `cannot use .* in struct literal` */ }
	_ = T1{c /* ERROR "unknown field" */ : 0}
	_ = T1{T0: { /* ERROR "missing type" */ }} // struct literal element type may not be elided
	_ = T1{T0: T0{}}
	_ = T1{T0 /* ERROR "invalid field name" */ .a: 0}

	// unkeyed elements
	_ = T0{1, 2, 3}
	_ = T0{1, b /* ERROR "mixture" */ : 2, 3}
	_ = T0{1, 2} /* ERROR "too few values" */
	_ = T0{1, 2, 3, 4  /* ERROR "too many values" */ }
	_ = T0{1, "foo" /* ERRORx `cannot use .* in struct literal` */, 3.4  /* ERRORx `cannot use .*\(truncated\)` */}

	// invalid type
	type P *struct{
		x int
	}
	_ = P /* ERROR "invalid composite literal type" */ {}

	// unexported fields
	_ = time.Time{}
	_ = time.Time{sec /* ERROR "unknown field" */ : 0}
	_ = time.Time{
		0 /* ERROR "implicit assignment to unexported field wall in struct literal" */,
		0 /* ERROR "implicit assignment" */ ,
		nil /* ERROR "implicit assignment" */ ,
	}
}

func array_literals() {
	type A0 [0]int
	_ = A0{}
	_ = A0{0 /* ERRORx `index .* out of bounds` */}
	_ = A0{0 /* ERRORx `index .* out of bounds` */ : 0}

	type A1 [10]int
	_ = A1{}
	_ = A1{0, 1, 2}
	_ = A1{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	_ = A1{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 /* ERRORx `index .* out of bounds` */ }
	_ = A1{- /* ERROR "negative" */ 1: 0}
	_ = A1{8: 8, 9}
	_ = A1{8: 8, 9, 10 /* ERRORx `index .* out of bounds` */ }
	_ = A1{0, 1, 2, 0 /* ERROR "duplicate index" */ : 0, 3: 3, 4}
	_ = A1{5: 5, 6, 7, 3: 3, 4}
	_ = A1{5: 5, 6, 7, 3: 3, 4, 5 /* ERROR "duplicate index" */ }
	_ = A1{10 /* ERRORx `index .* out of bounds` */ : 10, 10 /* ERRORx `index .* out of bounds` */ : 10}
	_ = A1{5: 5, 6, 7, 3: 3, 1 /* ERROR "overflows" */ <<100: 4, 5 /* ERROR "duplicate index" */ }
	_ = A1{5: 5, 6, 7, 4: 4, 1 /* ERROR "overflows" */ <<100: 4}
	_ = A1{2.0}
	_ = A1{2.1 /* ERROR "truncated" */ }
	_ = A1{"foo" /* ERRORx `cannot use .* in array or slice literal` */ }

	// indices must be integer constants
	i := 1
	const f = 2.1
	const s = "foo"
	_ = A1{i /* ERROR "index i must be integer constant" */ : 0}
	_ = A1{f /* ERROR "truncated" */ : 0}
	_ = A1{s /* ERROR "cannot convert" */ : 0}

	a0 := [...]int{}
	assert(len(a0) == 0)

	a1 := [...]int{0, 1, 2}
	assert(len(a1) == 3)
	var a13 [3]int
	var a14 [4]int
	a13 = a1
	a14 = a1 /* ERRORx `cannot use .* in assignment` */
	_, _ = a13, a14

	a2 := [...]int{- /* ERROR "negative" */ 1: 0}
	_ = a2

	a3 := [...]int{0, 1, 2, 0 /* ERROR "duplicate index" */ : 0, 3: 3, 4}
	assert(len(a3) == 5) // somewhat arbitrary

	a4 := [...]complex128{0, 1, 2, 1<<10-2: -1i, 1i, 400: 10, 12, 14}
	assert(len(a4) == 1024)

	// composite literal element types may be elided
	type T []int
	_ = [10]T{T{}, {}, 5: T{1, 2, 3}, 7: {1, 2, 3}}
	a6 := [...]T{T{}, {}, 5: T{1, 2, 3}, 7: {1, 2, 3}}
	assert(len(a6) == 8)

	// recursively so
	_ = [10][10]T{{}, [10]T{{}}, {{1, 2, 3}}}

	// from the spec
	type Point struct { x, y float32 }
	_ = [...]Point{Point{1.5, -3.5}, Point{0, 0}}
	_ = [...]Point{{1.5, -3.5}, {0, 0}}
	_ = [][]int{[]int{1, 2, 3}, []int{4, 5}}
	_ = [][]int{{1, 2, 3}, {4, 5}}
	_ = [...]*Point{&Point{1.5, -3.5}, &Point{0, 0}}
	_ = [...]*Point{{1.5, -3.5}, {0, 0}}
}

func slice_literals() {
	type S0 []int
	_ = S0{}
	_ = S0{0, 1, 2}
	_ = S0{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	_ = S0{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	_ = S0{- /* ERROR "negative" */ 1: 0}
	_ = S0{8: 8, 9}
	_ = S0{8: 8, 9, 10}
	_ = S0{0, 1, 2, 0 /* ERROR "duplicate index" */ : 0, 3: 3, 4}
	_ = S0{5: 5, 6, 7, 3: 3, 4}
	_ = S0{5: 5, 6, 7, 3: 3, 4, 5 /* ERROR "duplicate index" */ }
	_ = S0{10: 10, 10 /* ERROR "duplicate index" */ : 10}
	_ = S0{5: 5, 6, 7, 3: 3, 1 /* ERROR "overflows" */ <<100: 4, 5 /* ERROR "duplicate index" */ }
	_ = S0{5: 5, 6, 7, 4: 4, 1 /* ERROR "overflows" */ <<100: 4}
	_ = S0{2.0}
	_ = S0{2.1 /* ERROR "truncated" */ }
	_ = S0{"foo" /* ERRORx `cannot use .* in array or slice literal` */ }

	// indices must be resolved correctly
	const index1 = 1
	_ = S0{index1: 1}
	_ = S0{index2: 2}
	_ = S0{index3 /* ERROR "undefined" */ : 3}

	// indices must be integer constants
	i := 1
	const f = 2.1
	const s = "foo"
	_ = S0{i /* ERROR "index i must be integer constant" */ : 0}
	_ = S0{f /* ERROR "truncated" */ : 0}
	_ = S0{s /* ERROR "cannot convert" */ : 0}

	// composite literal element types may be elided
	type T []int
	_ = []T{T{}, {}, 5: T{1, 2, 3}, 7: {1, 2, 3}}
	_ = [][]int{{1, 2, 3}, {4, 5}}

	// recursively so
	_ = [][]T{{}, []T{{}}, {{1, 2, 3}}}

	// issue 17954
	type T0 *struct { s string }
	_ = []T0{{}}
	_ = []T0{{"foo"}}

	type T1 *struct{ int }
	_ = []T1{}
	_ = []T1{{0}, {1}, {2}}

	type T2 T1
	_ = []T2{}
	_ = []T2{{0}, {1}, {2}}

	_ = map[T0]T2{}
	_ = map[T0]T2{{}: {}}
}

const index2 int = 2

type N int
func (N) f() {}

func map_literals() {
	type M0 map[string]int
	type M1 map[bool]int
	type M2 map[*int]int

	_ = M0{}
	_ = M0{1 /* ERROR "missing key" */ }
	_ = M0{1 /* ERRORx `cannot use .* in map literal` */ : 2}
	_ = M0{"foo": "bar" /* ERRORx `cannot use .* in map literal` */ }
	_ = M0{"foo": 1, "bar": 2, "foo" /* ERROR "duplicate key" */ : 3 }

	_ = map[interface{}]int{2: 1, 2 /* ERROR "duplicate key" */ : 1}
	_ = map[interface{}]int{int(2): 1, int16(2): 1}
	_ = map[interface{}]int{int16(2): 1, int16 /* ERROR "duplicate key" */ (2): 1}

	type S string

	_ = map[interface{}]int{"a": 1, "a" /* ERROR "duplicate key" */ : 1}
	_ = map[interface{}]int{"a": 1, S("a"): 1}
	_ = map[interface{}]int{S("a"): 1, S /* ERROR "duplicate key" */ ("a"): 1}
	_ = map[interface{}]int{1.0: 1, 1.0 /* ERROR "duplicate key" */: 1}
	_ = map[interface{}]int{int64(-1): 1, int64 /* ERROR "duplicate key" */ (-1) : 1}
	_ = map[interface{}]int{^uint64(0): 1, ^ /* ERROR "duplicate key" */ uint64(0): 1}
	_ = map[interface{}]int{complex(1,2): 1, complex /* ERROR "duplicate key" */ (1,2) : 1}

	type I interface {
		f()
	}

	_ = map[I]int{N(0): 1, N(2): 1}
	_ = map[I]int{N(2): 1, N /* ERROR "duplicate key" */ (2): 1}

	// map keys must be resolved correctly
	key1 := "foo"
	_ = M0{key1: 1}
	_ = M0{key2: 2}
	_ = M0{key3 /* ERROR "undefined" */ : 2}

	var value int
	_ = M1{true: 1, false: 0}
	_ = M2{nil: 0, &value: 1}

	// composite literal element types may be elided
	type T [2]int
	_ = map[int]T{0: T{3, 4}, 1: {5, 6}}

	// recursively so
	_ = map[int][]T{0: {}, 1: {{}, T{1, 2}}}

	// composite literal key types may be elided
	_ = map[T]int{T{3, 4}: 0, {5, 6}: 1}

	// recursively so
	_ = map[[2]T]int{{}: 0, {{}}: 1, [2]T{{}}: 2, {T{1, 2}}: 3}

	// composite literal element and key types may be elided
	_ = map[T]T{{}: {}, {1, 2}: T{3, 4}, T{4, 5}: {}}
	_ = map[T]M0{{} : {}, T{1, 2}: M0{"foo": 0}, {1, 3}: {"foo": 1}}

	// recursively so
	_ = map[[2]T][]T{{}: {}, {{}}: {{}, T{1, 2}}, [2]T{{}}: nil, {T{1, 2}}: {{}, {}}}

	// from the spec
	type Point struct { x, y float32 }
	_ = map[string]Point{"orig": {0, 0}}
	_ = map[*Point]string{{0, 0}: "orig"}

	// issue 17954
	type T0 *struct{ s string }
	type T1 *struct{ int }
	type T2 T1

	_ = map[T0]T2{}
	_ = map[T0]T2{{}: {}}
}

var key2 string = "bar"

type I interface {
	m()
}

type I2 interface {
	m(int)
}

type T1 struct{}
type T2 struct{}

func (T2) m(int) {}

type mybool bool

func type_asserts() {
	var x int
	_ = x /* ERROR "not an interface" */ .(int)

	var e interface{}
	var ok bool
	x, ok = e.(int)
	_ = ok

	// ok value is of type bool
	var myok mybool
	_, myok = e.(int)
	_ = myok

	var t I
	_ = t /* ERRORx `use of .* outside type switch` */ .(type)
	_ = t /* ERROR "m has pointer receiver" */ .(T)
	_ = t.(*T)
	_ = t /* ERROR "missing method m" */ .(T1)
	_ = t /* ERROR "wrong type for method m" */ .(T2)
	_ = t /* STRICT "wrong type for method m" */ .(I2) // only an error in strict mode (issue 8561)

	// e doesn't statically have an m, but may have one dynamically.
	_ = e.(I2)
}

func f0() {}
func f1(x int) {}
func f2(u float32, s string) {}
func fs(s []byte) {}
func fv(x ...int) {}
func fi(x ... interface{}) {}
func (T) fm(x ...int)

func g0() {}
func g1() int { return 0}
func g2() (u float32, s string) { return }
func gs() []byte { return nil }

func _calls() {
	var x int
	var y float32
	var s []int

	f0()
	_ = f0 /* ERROR "used as value" */ ()
	f0(g0 /* ERROR "too many arguments" */ )

	f1(0)
	f1(x)
	f1(10.0)
	f1() /* ERROR "not enough arguments in call to f1\n\thave ()\n\twant (int)" */
	f1(x, y /* ERROR "too many arguments in call to f1\n\thave (int, float32)\n\twant (int)" */ )
	f1(s /* ERRORx `cannot use .* in argument` */ )
	f1(x ... /* ERROR "cannot use ..." */ )
	f1(g0 /* ERROR "used as value" */ ())
	f1(g1())
	f1(g2 /* ERROR "too many arguments in call to f1\n\thave (float32, string)\n\twant (int)" */ ())

	f2() /* ERROR "not enough arguments in call to f2\n\thave ()\n\twant (float32, string)" */
	f2(3.14) /* ERROR "not enough arguments in call to f2\n\thave (number)\n\twant (float32, string)" */
	f2(3.14, "foo")
	f2(x /* ERRORx `cannot use .* in argument` */ , "foo")
	f2(g0 /* ERROR "used as value" */ ()) /* ERROR "not enough arguments in call to f2\n\thave (func())\n\twant (float32, string)" */
	f2(g1()) /* ERROR "not enough arguments in call to f2\n\thave (int)\n\twant (float32, string)" */
	f2(g2())

	fs() /* ERROR "not enough arguments" */
	fs(g0 /* ERROR "used as value" */ ())
	fs(g1 /* ERRORx `cannot use .* in argument` */ ())
	fs(g2 /* ERROR "too many arguments" */ ())
	fs(gs())

	fv()
	fv(1, 2.0, x)
	fv(s /* ERRORx `cannot use .* in argument` */ )
	fv(s...)
	fv(x /* ERROR "cannot use" */ ...)
	fv(1, s /* ERROR "too many arguments" */ ...)
	fv(gs /* ERRORx `cannot use .* in argument` */ ())
	fv(gs /* ERRORx `cannot use .* in argument` */ ()...)

	var t T
	t.fm()
	t.fm(1, 2.0, x)
	t.fm(s /* ERRORx `cannot use .* in argument` */ )
	t.fm(g1())
	t.fm(1, s /* ERROR "too many arguments" */ ...)
	t.fm(gs /* ERRORx `cannot use .* in argument` */ ())
	t.fm(gs /* ERRORx `cannot use .* in argument` */ ()...)

	T.fm(t, )
	T.fm(t, 1, 2.0, x)
	T.fm(t, s /* ERRORx `cannot use .* in argument` */ )
	T.fm(t, g1())
	T.fm(t, 1, s /* ERROR "too many arguments" */ ...)
	T.fm(t, gs /* ERRORx `cannot use .* in argument` */ ())
	T.fm(t, gs /* ERRORx `cannot use .* in argument` */ ()...)

	var i interface{ fm(x ...int) } = t
	i.fm()
	i.fm(1, 2.0, x)
	i.fm(s /* ERRORx `cannot use .* in argument` */ )
	i.fm(g1())
	i.fm(1, s /* ERROR "too many arguments" */ ...)
	i.fm(gs /* ERRORx `cannot use .* in argument` */ ())
	i.fm(gs /* ERRORx `cannot use .* in argument` */ ()...)

	fi()
	fi(1, 2.0, x, 3.14, "foo")
	fi(g2())
	fi(0, g2)
	fi(0, g2 /* ERROR "multiple-value g2" */ ())
}

func issue6344() {
	type T []interface{}
	var x T
	fi(x...) // ... applies also to named slices
}

// GnoError:
// line 22: middle index required in 3-index slice
// line 23: middle index required in 3-index slice
// line 24: middle index required in 3-index slice
// line 90: middle index required in 3-index slice

// GoTypeCheckError:
// line 22

// UncaughtError:
// line 10: uncaught; gc expects: cannot index
// line 11: uncaught; gc expects: cannot index
// line 12: uncaught; gc expects: cannot slice
// line 15: uncaught; gc expects: cannot convert
// line 16: uncaught; gc expects: cannot convert
// line 17: uncaught; gc expects: truncated
// line 19: uncaught; gc expects: negative
// line 20: uncaught; gc expects: negative
// line 21: uncaught; gc expects: negative
// line 31: uncaught; gc expects: cannot use .* in assignment
// line 35: uncaught; gc expects: index .* out of bounds
// line 36: uncaught; gc expects: overflows
// line 37: uncaught; gc expects: constant shift overflow
// line 41: uncaught; gc expects: index .* out of bounds
// line 42: uncaught; gc expects: index .* out of bounds
// line 43: uncaught; gc expects: overflows
// line 45: uncaught; gc expects: index .* out of bounds
// line 46: uncaught; gc expects: index .* out of bounds
// line 47: uncaught; gc expects: invalid slice indices
// line 48: uncaught; gc expects: invalid slice indices
// line 49: uncaught; gc expects: invalid slice indices
// line 50: uncaught; gc expects: cannot take address
// line 54: uncaught; gc expects: index .* out of bounds
// line 55: uncaught; gc expects: overflows
// line 59: uncaught; gc expects: index .* out of bounds
// line 60: uncaught; gc expects: index .* out of bounds
// line 61: uncaught; gc expects: overflows
// line 63: uncaught; gc expects: index .* out of bounds
// line 64: uncaught; gc expects: index .* out of bounds
// line 65: uncaught; gc expects: invalid slice indices
// line 66: uncaught; gc expects: invalid slice indices
// line 67: uncaught; gc expects: invalid slice indices
// line 68: uncaught; gc expects: cannot take address
// line 71: uncaught; gc expects: index .* out of bounds
// line 77: uncaught; gc expects: index .* out of bounds
// line 80: uncaught; gc expects: negative
// line 81: uncaught; gc expects: negative
// line 82: uncaught; gc expects: negative
// line 85: uncaught; gc expects: invalid slice indices
// line 87: uncaught; gc expects: overflows
// line 88: uncaught; gc expects: overflows
// line 89: uncaught; gc expects: overflows | overflows
// line 92: uncaught; gc expects: invalid slice indices
// line 93: uncaught; gc expects: invalid slice indices
// line 94: uncaught; gc expects: invalid slice indices
// line 95: uncaught; gc expects: cannot take address
// line 98: uncaught; gc expects: cannot use .* in map index
// line 99: uncaught; gc expects: cannot slice
// line 106: uncaught; gc expects: mismatched types int and untyped string | cannot use 0
// line 109: uncaught; gc expects: negative
// line 110: uncaught; gc expects: negative
// line 111: uncaught; gc expects: negative
// line 112: uncaught; gc expects: 3-index slice of string
// line 113: uncaught; gc expects: 3-index slice of string
// line 118: uncaught; gc expects: cannot use .* in assignment
// line 121: uncaught; gc expects: index .* out of bounds
// line 124: uncaught; gc expects: negative
// line 125: uncaught; gc expects: negative
// line 126: uncaught; gc expects: negative
// line 131: uncaught; gc expects: cannot use .* in assignment
// line 132: uncaught; gc expects: index .* out of bounds
// line 133: uncaught; gc expects: index .* out of bounds
// line 145: uncaught; gc expects: cannot use .* in assignment
// line 146: uncaught; gc expects: cannot use .* in assignment
// line 158: uncaught; gc expects: no field or method
// line 159: uncaught; gc expects: has no method
// line 160: uncaught; gc expects: invalid method expression T.m (needs pointer receiver (*T).m)
// line 163: uncaught; gc expects: invalid method expression T.m (needs pointer receiver (*T).m)
// line 167: uncaught; gc expects: has no method
// line 168: uncaught; gc expects: has no method
// line 185: uncaught; gc expects: mixture of .* elements
// line 186: uncaught; gc expects: unknown field
// line 187: uncaught; gc expects: invalid field name
// line 188: uncaught; gc expects: duplicate field
// line 189: uncaught; gc expects: cannot use .* in struct literal
// line 190: uncaught; gc expects: unknown field
// line 191: uncaught; gc expects: missing type
// line 193: uncaught; gc expects: invalid field name
// line 197: uncaught; gc expects: mixture
// line 198: uncaught; gc expects: too few values
// line 199: uncaught; gc expects: too many values
// line 200: uncaught; gc expects: cannot use .* in struct literal | cannot use .*\(truncated\)
// line 206: uncaught; gc expects: invalid composite literal type
// line 210: uncaught; gc expects: unknown field
// line 212: uncaught; gc expects: implicit assignment to unexported field wall in struct literal
// line 213: uncaught; gc expects: implicit assignment
// line 214: uncaught; gc expects: implicit assignment
// line 221: uncaught; gc expects: index .* out of bounds
// line 222: uncaught; gc expects: index .* out of bounds
// line 228: uncaught; gc expects: index .* out of bounds
// line 229: uncaught; gc expects: negative
// line 231: uncaught; gc expects: index .* out of bounds
// line 232: uncaught; gc expects: duplicate index
// line 234: uncaught; gc expects: duplicate index
// line 235: uncaught; gc expects: index .* out of bounds | index .* out of bounds
// line 236: uncaught; gc expects: overflows | duplicate index
// line 237: uncaught; gc expects: overflows
// line 239: uncaught; gc expects: truncated
// line 240: uncaught; gc expects: cannot use .* in array or slice literal
// line 246: uncaught; gc expects: index i must be integer constant
// line 247: uncaught; gc expects: truncated
// line 248: uncaught; gc expects: cannot convert
// line 258: uncaught; gc expects: cannot use .* in assignment
// line 261: uncaught; gc expects: negative
// line 264: uncaught; gc expects: duplicate index
// line 295: uncaught; gc expects: negative
// line 298: uncaught; gc expects: duplicate index
// line 300: uncaught; gc expects: duplicate index
// line 301: uncaught; gc expects: duplicate index
// line 302: uncaught; gc expects: overflows | duplicate index
// line 303: uncaught; gc expects: overflows
// line 305: uncaught; gc expects: truncated
// line 306: uncaught; gc expects: cannot use .* in array or slice literal
// line 312: uncaught; gc expects: undefined
// line 318: uncaught; gc expects: index i must be integer constant
// line 319: uncaught; gc expects: truncated
// line 320: uncaught; gc expects: cannot convert
// line 358: uncaught; gc expects: missing key
// line 359: uncaught; gc expects: cannot use .* in map literal
// line 360: uncaught; gc expects: cannot use .* in map literal
// line 361: uncaught; gc expects: duplicate key
// line 363: uncaught; gc expects: duplicate key
// line 365: uncaught; gc expects: duplicate key
// line 369: uncaught; gc expects: duplicate key
// line 371: uncaught; gc expects: duplicate key
// line 372: uncaught; gc expects: duplicate key
// line 373: uncaught; gc expects: duplicate key
// line 374: uncaught; gc expects: duplicate key
// line 375: uncaught; gc expects: duplicate key
// line 382: uncaught; gc expects: duplicate key
// line 388: uncaught; gc expects: undefined
// line 447: uncaught; gc expects: not an interface
// line 460: uncaught; gc expects: use of .* outside type switch
// line 461: uncaught; gc expects: m has pointer receiver
// line 463: uncaught; gc expects: missing method m
// line 464: uncaught; gc expects: wrong type for method m
// line 490: uncaught; gc expects: used as value
// line 491: uncaught; gc expects: too many arguments
// line 496: uncaught; gc expects: not enough arguments in call to f1
// 	have ()
// 	want (int)
// line 497: uncaught; gc expects: too many arguments in call to f1
// 	have (int, float32)
// 	want (int)
// line 498: uncaught; gc expects: cannot use .* in argument
// line 499: uncaught; gc expects: cannot use ...
// line 500: uncaught; gc expects: used as value
// line 502: uncaught; gc expects: too many arguments in call to f1
// 	have (float32, string)
// 	want (int)
// line 504: uncaught; gc expects: not enough arguments in call to f2
// 	have ()
// 	want (float32, string)
// line 505: uncaught; gc expects: not enough arguments in call to f2
// 	have (number)
// 	want (float32, string)
// line 507: uncaught; gc expects: cannot use .* in argument
// line 508: uncaught; gc expects: used as value | not enough arguments in call to f2
// 	have (func())
// 	want (float32, string)
// line 509: uncaught; gc expects: not enough arguments in call to f2
// 	have (int)
// 	want (float32, string)
// line 512: uncaught; gc expects: not enough arguments
// line 513: uncaught; gc expects: used as value
// line 514: uncaught; gc expects: cannot use .* in argument
// line 515: uncaught; gc expects: too many arguments
// line 520: uncaught; gc expects: cannot use .* in argument
// line 522: uncaught; gc expects: cannot use
// line 523: uncaught; gc expects: too many arguments
// line 524: uncaught; gc expects: cannot use .* in argument
// line 525: uncaught; gc expects: cannot use .* in argument
// line 530: uncaught; gc expects: cannot use .* in argument
// line 532: uncaught; gc expects: too many arguments
// line 533: uncaught; gc expects: cannot use .* in argument
// line 534: uncaught; gc expects: cannot use .* in argument
// line 538: uncaught; gc expects: cannot use .* in argument
// line 540: uncaught; gc expects: too many arguments
// line 541: uncaught; gc expects: cannot use .* in argument
// line 542: uncaught; gc expects: cannot use .* in argument
// line 547: uncaught; gc expects: cannot use .* in argument
// line 549: uncaught; gc expects: too many arguments
// line 550: uncaught; gc expects: cannot use .* in argument
// line 551: uncaught; gc expects: cannot use .* in argument
// line 557: uncaught; gc expects: multiple-value g2

// KnownIssue:
// errorcheck: 0 over-strict line(s), 171 uncaught marker(s)
