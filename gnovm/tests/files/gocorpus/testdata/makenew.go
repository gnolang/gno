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
// line 14: invalid argument: cannot make int; type must be slice, map
// line 15: invalid operation: make([]int) expects 2 or 3 arguments; found 1
// line 17: wrong argument count in call to (const (new func(<T.(type)>{}) <*T>{})); want 1 got 0
// line 18: wrong argument count in call to (const (new func(<T.(type)>{}) <*T>{})); want 1 got 2

// GnoWordingDiffers:
// line 17: gc expects: missing argument|not enough arguments; gno: wrong argument count in call to (const (new func(<T.(type)>{}) <*T>{})); want 1 got 0
// line 18: gc expects: too many arguments; gno: wrong argument count in call to (const (new func(<T.(type)>{}) <*T>{})); want 1 got 2

// Verdict: Clean errorcheck
