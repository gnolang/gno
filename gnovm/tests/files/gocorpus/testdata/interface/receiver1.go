// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify compiler complains about missing implicit methods.
// Does not compile.

package main

type T int

func (t T) V()
func (t *T) P()

type V interface {
	V()
}
type P interface {
	P()
	V()
}

type S struct {
	T
}
type SP struct {
	*T
}

func main() {
	var t T
	var v V
	var p P
	var s S
	var sp SP

	v = t
	p = t // ERROR "does not implement|requires a pointer|cannot use"
	_, _ = v, p
	v = &t
	p = &t
	_, _ = v, p

	v = s
	p = s // ERROR "does not implement|requires a pointer|cannot use"
	_, _ = v, p
	v = &s
	p = &s
	_, _ = v, p

	v = sp
	p = sp // no error!
	_, _ = v, p
	v = &sp
	p = &sp
	_, _ = v, p
}

// GnoError:
// line 40: main.T does not implement main.P (method P has pointer receiver)
// line 47: main.S does not implement main.P (method P has pointer receiver)

// GoTypeCheckError:
// line 40: cannot use t (variable of int type T) as P value in assignment: T does not implement P (method P has pointer receiver)
// line 47: cannot use s (variable of struct type S) as P value in assignment: S does not implement P (method P has pointer receiver)
