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
// line 26: StaticBlock.Define2(f) cannot change .T
// line 32: StaticBlock.Define2(i) cannot change .T
// line 38: no new variables on left side of := (and 4 more errors)
// line 44: no new variables on left side of := (and 3 more errors)
// line 45: expected 0 return values
// line 49: expected declaration, found '{'
// line 51: expected declaration, found i
// line 52: expected declaration, found i
// line 53: expected declaration, found _
// line 54: expected declaration, found '}'
// line 56: expected declaration, found '{'
// line 57: expected declaration, found i
// line 58: expected declaration, found i
// line 59: expected declaration, found _
// line 60: expected declaration, found '}'
// line 61: expected declaration, found '{'
// line 63: expected declaration, found i
// line 64: expected declaration, found i
// line 65: expected declaration, found _
// line 66: expected declaration, found '}'
// line 67: expected declaration, found '}'

// GoTypeCheckError:
// line 20: no new variables on left side of :=
// line 26: cannot use f3() (value of type float32) as int value in assignment
// line 32: cannot use f3() (value of type int) as float32 value in assignment
// line 38: no new variables on left side of :=
// line 44: no new variables on left side of :=
// line 52: no new variables on left side of :=
// line 58: no new variables on left side of :=
// line 64: no new variables on left side of :=

// GnoOverStrictError:
// line 45: expected 0 return values
// line 49: expected declaration, found '{'
// line 51: expected declaration, found i
// line 53: expected declaration, found _
// line 54: expected declaration, found '}'
// line 56: expected declaration, found '{'
// line 57: expected declaration, found i
// line 59: expected declaration, found _
// line 60: expected declaration, found '}'
// line 61: expected declaration, found '{'
// line 63: expected declaration, found i
// line 65: expected declaration, found _
// line 66: expected declaration, found '}'
// line 67: expected declaration, found '}'
