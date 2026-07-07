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
// line 15: 2: name b not defined in fileset with files [typecheck.go]
// line 16: expected declaration, found 'return'
// line 17: expected declaration, found '}'

// GoTypeCheckError:
// line 15: undefined: b
// line 16: undefined: b
// line 20: not enough arguments in call to mine
// 	have ()
// 	want (unknown type)
// line 21: not enough arguments in call to mine
// 	have ()
// 	want (unknown type)

// GnoOverStrictError:
// line 17: expected declaration, found '}'
