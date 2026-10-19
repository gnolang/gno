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
// line 24: impossible type switch case: x cannot have dynamic type string (missing method M)
// line 39: no new variables on left side of :=
// line 45: cannot type switch on non-interface value i (type int)
// line 54: cannot type switch on non-interface value s (type main[main/typeswitch3.go:43:1-56:2].S)

// Verdict: Clean errorcheck
