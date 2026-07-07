// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func foo() (T, T) { // ERROR "undefined"
	return 0, 0
}

func bar() (T, string, T) { // ERROR "undefined"
	return 0, "", 0
}

func main() {
	var x, y, z int
	x, y = foo()
	x, y, z = bar() // ERROR "cannot (use type|assign|use.*type) string|"
	_, _, _ = x, y, z
}

// GnoError:
// line 9: 2: name T not defined in fileset with files [issue6572.go]
// line 10: expected declaration, found 'return' (and 1 more errors)
// line 11: expected declaration, found '}' (and 1 more errors)
// line 13: 2: name T not defined in fileset with files [issue6572.go]
// line 14: expected declaration, found 'return' (and 1 more errors)
// line 15: expected declaration, found '}' (and 1 more errors)

// GoTypeCheckError:
// line 9: undefined: T
// line 13: undefined: T
// line 20: cannot use bar() (value of type string) as int value in assignment

// GnoOverStrictError:
// line 10: expected declaration, found 'return' (and 1 more errors)
// line 11: expected declaration, found '}' (and 1 more errors)
// line 14: expected declaration, found 'return' (and 1 more errors)
// line 15: expected declaration, found '}' (and 1 more errors)
