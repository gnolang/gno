// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f1() {
	for a, a := range []int{1, 2, 3} { // ERROR "a.* repeated on left side of :=|a redeclared"
		println(a)
	}
}

func f2() {
	var a int
	for a, a := range []int{1, 2, 3} { // ERROR "a.* repeated on left side of :=|a redeclared"
		println(a)
	}
	println(a)
}

// GnoIncomplete: covered 1 of 2 markers; Gno bailed before the rest — a runnable variant is needed to exercise them
// GnoError:
// line 10: a redeclared in this block
