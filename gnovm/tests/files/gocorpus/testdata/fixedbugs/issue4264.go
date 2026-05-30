// errorcheck

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 4264: reject int division by const 0

package main

func main() {
	var x int
	var y float64
	var z complex128

	println(x/0) // ERROR "division by zero"
	println(y/0)
	println(z/0)
}

// GoTypeCheckError:
// line 16: invalid operation: division by zero

// KnownIssue:
// line 14: name complex128 not defined in fileset with files [issue4264.go]
