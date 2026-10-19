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
// line 15: name not defined: b
// line 16: name b not declared
// line 20: wrong argument count in call to mine<VPBlock(3,0)>; want 1 got 0
// line 21: wrong argument count in call to mine<VPBlock(3,0)>; want 1 got 0

// GnoWordingDiffers:
// line 15: gc expects: undefined.*b; gno: name not defined: b
// line 16: gc expects: undefined.*b; gno: name b not declared
// line 20: gc expects: not enough arguments; gno: wrong argument count in call to mine<VPBlock(3,0)>; want 1 got 0
// line 21: gc expects: undefined.*c|not enough arguments; gno: wrong argument count in call to mine<VPBlock(3,0)>; want 1 got 0

// Verdict: Clean errorcheck
