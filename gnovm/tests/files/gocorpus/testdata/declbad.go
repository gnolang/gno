// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that incorrect short declarations and redeclarations are detected.
// Does not compile.

package main

func f1() int                    { return 1 }
func f2() (float32, int)         { return 1, 2 }
func f3() (float32, int, string) { return 1, 2, "3" }

func main() {
	{
		// simple redeclaration
		i := f1()
		i := f1() // ERROR "redeclared|no new"
		_ = i
	}
	{
		// change of type for f
		i, f, s := f3()
		f, g, t := f3() // ERROR "redeclared|cannot assign|incompatible|cannot use"
		_, _, _, _, _ = i, f, s, g, t
	}
	{
		// change of type for i
		i, f, s := f3()
		j, i, t := f3() // ERROR "redeclared|cannot assign|incompatible|cannot use"
		_, _, _, _, _ = i, f, s, j, t
	}
	{
		// no new variables
		i, f, s := f3()
		i, f := f2() // ERROR "redeclared|no new"
		_, _, _ = i, f, s
	}
	{
		// multiline no new variables
		i := f1
		i := func() int { // ERROR "redeclared|no new|incompatible"
			return 0
		}
		_ = i
	}
	{
		// single redeclaration
		i, f, s := f3()
		i := 1 // ERROR "redeclared|no new|incompatible"
		_, _, _ = i, f, s
	}
	// double redeclaration
	{
		i, f, s := f3()
		i, f := f2() // ERROR "redeclared|no new"
		_, _, _ = i, f, s
	}
	{
		// triple redeclaration
		i, f, s := f3()
		i, f, s := f3() // ERROR "redeclared|no new"
		_, _, _ = i, f, s
	}
}

// GnoError:
// line 20: no new variables on left side of := (and 5 more errors)

// GoTypeCheckError:
// line 20
// line 26
// line 32
// line 38
// line 44
// line 52
// line 58
// line 64

// GoTypesOnly:
// line 26: gc expects: redeclared|cannot assign|incompatible|cannot use
// line 32: gc expects: redeclared|cannot assign|incompatible|cannot use
// line 38: gc expects: redeclared|no new
// line 44: gc expects: redeclared|no new|incompatible
// line 52: gc expects: redeclared|no new|incompatible
// line 58: gc expects: redeclared|no new
// line 64: gc expects: redeclared|no new

// Verdict: Clean errorcheck
