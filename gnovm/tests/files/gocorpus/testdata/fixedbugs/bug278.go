// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is a test case for issue 804.

package main

func f() [10]int {
	return [10]int{}
}

var m map[int][10]int

func main() {
	f()[1] = 2	// ERROR "cannot|invalid"
	f()[2:3][0] = 4	// ERROR "cannot|addressable"
	var x = "abc"
	x[2] = 3	// ERROR "cannot|invalid"
	m[0][5] = 6  // ERROR "cannot|invalid"
}

// GnoError:
// line 21: cannot assign to x<VPBlock(1,0)>[(const (2 int))]

// GoTypeCheckError:
// line 18
// line 19
// line 21
// line 22

// GoTypesOnly:
// line 18: gc expects: cannot|invalid
// line 19: gc expects: cannot|addressable
// line 22: gc expects: cannot|invalid

// Verdict: Clean errorcheck
