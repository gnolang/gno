// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify compiler messages about erroneous static interface conversions.
// Does not compile.

package main

type T struct {
	a int
}

var t *T

type X int

func (x *X) M() {}

type I interface {
	M()
}

var i I

type I2 interface {
	M()
	N()
}

var i2 I2

type E interface{}

var e E

func main() {
	e = t // ok
	t = e // ERROR "need explicit|need type assertion"

	// neither of these can work,
	// because i has an extra method
	// that t does not, so i cannot contain a t.
	i = t // ERROR "incompatible|missing method M"
	t = i // ERROR "incompatible|assignment$"

	i = i2 // ok
	i2 = i // ERROR "incompatible|missing method N"

	i = I(i2)  // ok
	i2 = I2(i) // ERROR "invalid|missing N method|cannot convert"

	e = E(t) // ok
	t = T(e) // ERROR "need explicit|need type assertion|incompatible|cannot convert"

	// cannot type-assert non-interfaces
	f := 2.0
	_ = f.(int) // ERROR "non-interface type|only valid for interface types|not an interface"

}

type M interface {
	M()
}

var m M

var _ = m.(int) // ERROR "impossible type assertion"

type Int int

func (Int) M(float64) {}

var _ = m.(Int) // ERROR "impossible type assertion"

var _ = m.(X) // ERROR "pointer receiver"

var ii int
var jj Int

var m1 M = ii // ERROR "incompatible|missing"
var m2 M = jj // ERROR "incompatible|wrong type for method M"

var m3 = M(ii) // ERROR "invalid|missing|cannot convert"
var m4 = M(jj) // ERROR "invalid|wrong type for M method|cannot convert"

type B1 interface {
	_() // ERROR "methods must have a unique non-blank name"
}

type B2 interface {
	M()
	_() // ERROR "methods must have a unique non-blank name"
}

type T2 struct{}

func (t *T2) M() {}
func (t *T2) _() {}

// Already reported about the invalid blank interface method above;
// no need to report about not implementing it.
var b1 B1 = &T2{}
var b2 B2 = &T2{}

// GnoError:
// line 41: cannot use interface {} as *main.T
// line 46: *main.T does not implement main.I (missing method M)
// line 47: cannot use interface {M func()} as *main.T
// line 50: main.I does not implement main.I2 (missing method N)
// line 53: main.I does not implement main.I2 (missing method N)
// line 56: cannot convert main.E to main.T: need type assertion
// line 60: invalid operation: f<VPBlock(1,0)> (variable of type float64) is not an interface
// line 83: int does not implement main.M (missing method M)
// line 84: main.Int does not implement main.M (wrong type for method M)
// line 86: int does not implement main.M (missing method M)
// line 87: main.Int does not implement main.M (wrong type for method M)

// GoTypeCheckError:
// line 41: cannot use e (variable of interface type E) as *T value in assignment: need type assertion
// line 46: cannot use t (variable of type *T) as I value in assignment: *T does not implement I (missing method M)
// line 47: cannot use i (variable of interface type I) as *T value in assignment
// line 50: cannot use i (variable of interface type I) as I2 value in assignment: I does not implement I2 (missing method N)
// line 53: cannot convert i (variable of interface type I) to type I2: I does not implement I2 (missing method N)
// line 56: cannot convert e (variable of interface type E) to type T: need type assertion
// line 60: invalid operation: f (variable of type float64) is not an interface
// line 70: impossible type assertion: m.(int)
// 	int does not implement M (missing method M)
// line 76: impossible type assertion: m.(Int)
// 	Int does not implement M (wrong type for method M)
// 		have M(float64)
// 		want M()
// line 78: impossible type assertion: m.(X)
// 	X does not implement M (method M has pointer receiver)
// line 83: cannot use ii (variable of type int) as M value in variable declaration: int does not implement M (missing method M)
// line 84: cannot use jj (variable of int type Int) as M value in variable declaration: Int does not implement M (wrong type for method M)
// 		have M(float64)
// 		want M()
// line 86: cannot convert ii (variable of type int) to type M: int does not implement M (missing method M)
// line 87: cannot convert jj (variable of int type Int) to type M: Int does not implement M (wrong type for method M)
// 		have M(float64)
// 		want M()
// line 90: methods must have a unique non-blank name
// line 95: methods must have a unique non-blank name
