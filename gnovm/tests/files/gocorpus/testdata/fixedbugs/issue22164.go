// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test error recovery after missing closing parentheses in lists.

package p

func f() {
	x := f(g() // ERROR "unexpected newline"
	y := 1
}

func g() {
}

func h() {
	x := f(g() // ERROR "unexpected newline"
}

func i() {
	x := []int{1, 2, 3 // ERROR "unexpected newline"
	y := 0
}

// GnoError:
// line 12: missing ',' before newline in argument list

// GoTypeCheckError:
// line 12

// ParseBlocked:
// line 20: gc expects: unexpected newline
// line 24: gc expects: unexpected newline

// GnoWordingDiffers:
// line 12: gc expects: unexpected newline; gno: missing ',' before newline in argument list

// Verdict: Clean errorcheck
