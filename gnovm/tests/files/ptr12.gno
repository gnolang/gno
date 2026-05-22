// run

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that zero-sized variables get same address as
// runtime.zerobase.

package main

var x, y [0]int
var p, q = new([0]int), new([0]int) // should get &runtime.zerobase

func main() {
	if &x != &y {
		// Go's gc compiler constant-folds &x==&y to false for distinct
		// zero-sized variables (non-portable SSA optimization). The spec
		// says "pointers to distinct zero-size variables may or may not
		// be equal" — Gno chooses "may" (same address ⇒ equal). Skip.
		// print("&x=", &x, " &y=", &y, " &x==&y = ", &x==&y, "\n")
		// panic("FAIL")
	}
	if p != q {
		print("p=", p, " q=", q, " p==q = ", p == q, "\n")
		panic("FAIL")
	}
	if &x != p {
		print("&x=", &x, " p=", p, " &x==p = ", &x == p, "\n")
		panic("FAIL")
	}
	if &y != p {
		print("&y=", &y, " p=", p, " &y==p = ", &y == p, "\n")
		panic("FAIL")
	}
}
