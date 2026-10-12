// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T struct {
	x, y *T
}

func main() {
	// legal composite literals
	_ = struct{}{}
	_ = [42]int{}
	_ = [...]int{}
	_ = []int{}
	_ = map[int]int{}
	_ = T{}

	// illegal composite literals: parentheses not allowed around literal type
	_ = (struct{}){}    // ERROR "parenthesize"
	_ = ([42]int){}     // ERROR "parenthesize"
	_ = ([...]int){}    // ERROR "parenthesize"
	_ = ([]int){}       // ERROR "parenthesize"
	_ = (map[int]int){} // ERROR "parenthesize"
	_ = (T){}           // ERROR "parenthesize"
}

// GnoError:
// line 23: cannot parenthesize type in composite literal (and 5 more errors)

// GoTypeCheckError:
// line 23

// UncaughtError:
// line 24: uncaught; gc expects: parenthesize
// line 25: uncaught; gc expects: parenthesize
// line 26: uncaught; gc expects: parenthesize
// line 27: uncaught; gc expects: parenthesize
// line 28: uncaught; gc expects: parenthesize

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 5 uncaught marker(s)
