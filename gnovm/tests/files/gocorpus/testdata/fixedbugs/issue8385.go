// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 8385: provide a more descriptive error when a method expression
// is called without a receiver.

package main

type Fooer interface {
	Foo(i, j int)
}

func f(x int) {
}

type I interface {
	M(int)
}
type T struct{}

func (t T) M(x int) {
}

func g() func(int)

func main() {
	Fooer.Foo(5, 6) // ERROR "not enough arguments in call to method expression Fooer.Foo|incompatible type|not enough arguments"

	var i I
	var t *T

	g()()    // ERROR "not enough arguments in call to g\(\)|not enough arguments"
	f()      // ERROR "not enough arguments in call to f|not enough arguments"
	i.M()    // ERROR "not enough arguments in call to i\.M|not enough arguments"
	I.M()    // ERROR "not enough arguments in call to method expression I\.M|not enough arguments"
	t.M()    // ERROR "not enough arguments in call to t\.M|not enough arguments"
	T.M()    // ERROR "not enough arguments in call to method expression T\.M|not enough arguments"
	(*T).M() // ERROR "not enough arguments in call to method expression \(\*T\)\.M|not enough arguments"
}

// GoTypeCheckError:
// line 30: not enough arguments in call to Fooer.Foo
// 	have (number, number)
// 	want (Fooer, int, int)
// line 35: not enough arguments in call to g()
// 	have ()
// 	want (int)
// line 36: not enough arguments in call to f
// 	have ()
// 	want (int)
// line 37: not enough arguments in call to i.M
// 	have ()
// 	want (int)
// line 38: not enough arguments in call to I.M
// 	have ()
// 	want (I, int)
// line 39: not enough arguments in call to t.M
// 	have ()
// 	want (int)
// line 40: not enough arguments in call to T.M
// 	have ()
// 	want (T, int)
// line 41: not enough arguments in call to (*T).M
// 	have ()
// 	want (*T, int)

// KnownIssue:
// line 27: function g does not have a body but is not natively defined (did you build after pulling from the repository?)
