// errorcheck -0 -m

// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test no inlining of function calls in testing.B.Loop.
// See issue #61515.

package foo

import "testing"

func caninline(x int) int { // ERROR "can inline caninline"
	return x
}

func test(b *testing.B) { // ERROR "leaking param: b"
	for i := 0; i < b.N; i++ {
		caninline(1) // ERROR "inlining call to caninline"
	}
	for b.Loop() { // ERROR "skip inlining within testing.B.loop" "inlining call to testing\.\(\*B\)\.Loop"
		caninline(1)
	}
	for i := 0; i < b.N; i++ {
		caninline(1) // ERROR "inlining call to caninline"
	}
	for b.Loop() { // ERROR "skip inlining within testing.B.loop" "inlining call to testing\.\(\*B\)\.Loop"
		caninline(1)
	}
	for i := 0; i < b.N; i++ {
		caninline(1) // ERROR "inlining call to caninline"
	}
	for b.Loop() { // ERROR "skip inlining within testing.B.loop" "inlining call to testing\.\(\*B\)\.Loop"
		caninline(1)
	}
}

// GnoIncomplete: covered 3 of 8 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// GnoError:
// line 22: missing field Loop in *testing.B

// GoTypeCheckError:
// line 22: b.Loop undefined (type *testing.B has no field or method Loop)
// line 28: b.Loop undefined (type *testing.B has no field or method Loop)
// line 34: b.Loop undefined (type *testing.B has no field or method Loop)
