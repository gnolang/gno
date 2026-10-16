// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that the Go compiler will not
// die after running into an undefined
// type in the argument list for a
// function.
// Does not compile.

package main

func mine(int b) int { // ERROR "undefined.*b"
	return b + 2 // ERROR "undefined.*b"
}

func main() {
	mine()     // ERROR "not enough arguments"
	c = mine() // ERROR "undefined.*c|not enough arguments"
}

// GnoError:
// line 15: name b not defined in fileset with files [typecheck.go]

// GnoWordingDiffers:
// line 15: gc expects: undefined.*b; gno: 2: name b not defined in fileset with files [typecheck.go]
// line 16: gc expects: undefined.*b; gno: 2: name b not defined in fileset with files [typecheck.go]

// Verdict: KnownIssue errorcheck: #6344; uncaught markers: lines 20, 21
