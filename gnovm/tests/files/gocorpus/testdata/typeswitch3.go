// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that erroneous type switches are caught by the compiler.
// Issue 2700, among other things.
// Does not compile.

package main

import (
	"io"
)

type I interface {
	M()
}

func main() {
	var x I
	switch x.(type) {
	case string: // ERROR "impossible"
		println("FAIL")
	}

	// Issue 2700: if the case type is an interface, nothing is impossible

	var r io.Reader

	_, _ = r.(io.Writer)

	switch r.(type) {
	case io.Writer:
	}

	// Issue 2827.
	switch _ := r.(type) { // ERROR "invalid variable name _|no new variables?"
	}
}

func noninterface() {
	var i int
	switch i.(type) { // ERROR "cannot type switch on non-interface value|not an interface"
	case string:
	case int:
	}

	type S struct {
		name string
	}
	var s S
	switch s.(type) { // ERROR "cannot type switch on non-interface value|not an interface"
	}
}

// GnoError:
// line 39: no new variables on left side of :=

// GoTypeCheckError:
// line 24: impossible type switch case: string
// 	x (variable of interface type I) cannot have dynamic type string (missing method M)
// line 39: no new variable on left side of :=
// line 45: i (variable of type int) is not an interface
// line 54: s (variable of struct type S) is not an interface
