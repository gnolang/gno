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
// line 9: name T not defined in fileset with files [issue6572.go]
// line 13: name T not defined in fileset with files [issue6572.go]

// GnoWordingDiffers:
// line 9: gc expects: undefined; gno: name T not defined in fileset with files [issue6572.go]
// line 13: gc expects: undefined; gno: name T not defined in fileset with files [issue6572.go]

// Verdict: KnownIssue errorcheck: #6344; uncaught markers: lines 20
