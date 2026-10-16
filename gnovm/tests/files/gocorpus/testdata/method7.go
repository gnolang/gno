// run

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test forms of method expressions T.m where T is
// a literal type.

package main

var got, want string

type I interface {
	m()
}

type S struct {
}

func (S) m()          { got += " m()" }
func (S) m1(s string) { got += " m1(" + s + ")" }

type T int

func (T) m2() { got += " m2()" }

type Outer struct{ *Inner }
type Inner struct{ s string }

func (i Inner) M() string { return i.s }

func main() {
	// method expressions with named receiver types
	I.m(S{})
	want += " m()"

	S.m1(S{}, "a")
	want += " m1(a)"

	// method expressions with literal receiver types
	f := interface{ m1(string) }.m1
	f(S{}, "b")
	want += " m1(b)"

	interface{ m1(string) }.m1(S{}, "c")
	want += " m1(c)"

	x := S{}
	interface{ m1(string) }.m1(x, "d")
	want += " m1(d)"

	g := struct{ T }.m2
	g(struct{ T }{})
	want += " m2()"

	if got != want {
		panic("got" + got + ", want" + want)
	}

	h := (*Outer).M
	got := h(&Outer{&Inner{"hello"}})
	want := "hello"
	if got != want {
		panic("got " + got + ", want " + want)
	}
}


// Tracked: issue #5787 (method expressions: interface/promoted/mixed-receiver forms); broken on master, no PR yet.

// GnoOutput:

// GnoError:
// main/method7.go:42:7-33: invalid method expression interface { m1 func(.arg_0 (const-type string)) }.m1 (type interface {m1 func(string)} has no method m1)

// GoOutput:

// Verdict: KnownIssue run: #5787 method expression on an interface or mixed-receiver type is rejected at preprocess
