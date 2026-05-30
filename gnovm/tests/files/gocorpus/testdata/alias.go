// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that error messages say what the source file says
// (uint8 vs byte, int32 vs. rune).
// Does not compile.

package main

import (
	"fmt"
	"unicode/utf8"
)

func f(byte)  {}
func g(uint8) {}

func main() {
	var x float64
	f(x) // ERROR "byte"
	g(x) // ERROR "uint8"

	// Test across imports.

	var ff fmt.Formatter
	var fs fmt.State
	ff.Format(fs, x) // ERROR "rune"

	utf8.RuneStart(x) // ERROR "byte"
}

// GnoError:
// line 23: cannot use float64 as uint8
// line 24: cannot use float64 as uint8
// line 30: cannot use float64 as int32
// line 32: cannot use float64 as uint8

// GoTypeCheckError:
// line 23: cannot use x (variable of type float64) as byte value in argument to f
// line 24: cannot use x (variable of type float64) as uint8 value in argument to g
// line 30: cannot use x (variable of type float64) as rune value in argument to ff.Format
// line 32: cannot use x (variable of type float64) as byte value in argument to utf8.RuneStart
