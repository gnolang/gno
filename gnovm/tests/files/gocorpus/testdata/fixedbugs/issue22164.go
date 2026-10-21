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
// line 12: unexpected newline in argument list; possibly missing comma or )
// line 20: unexpected newline in argument list; possibly missing comma or )
// line 24: unexpected newline in composite literal; possibly missing comma or }
// line 26: expected '(', found main

// Verdict: KnownIssue errorcheck: #6344: line 26 is the harness's appended main parsed inside i, whose '}' the comma recovery took as the literal's close (gc hides the same follow-up at EOF); over-strict: lines 26
