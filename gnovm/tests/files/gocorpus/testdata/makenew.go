// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that make and new arguments requirements are enforced by the
// compiler.

package main

func main() {
	_ = make()      // ERROR "missing argument|not enough arguments"
	_ = make(int)   // ERROR "cannot make type|cannot make int"
	_ = make([]int) // ERROR "missing len argument|expects 2 or 3 arguments"

	_ = new()       // ERROR "missing argument|not enough arguments"
	_ = new(int, 2) // ERROR "too many arguments"
}

// GnoError:
// line 13: missing argument to make
// line 14: invalid argument: cannot make int
// line 15: invalid operation: make([]int) expects 2 or 3 arguments
// line 17: wrong argument count in call to (const (new func(<T.(type)>{}) <*T>{}))
// line 18: wrong argument count in call to (const (new func(<T.(type)>{}) <*T>{}))

// GoTypeCheckError:
// line 13: invalid operation: not enough arguments for make() (expected 1, found 0)
// line 14: invalid argument: cannot make int: type must be slice, map, or channel
// line 15: invalid operation: make([]int) expects 2 or 3 arguments
// line 17: invalid operation: not enough arguments for new() (expected 1, found 0)
// line 18: invalid operation: too many arguments for new(int, 2) (expected 1, found 2)
