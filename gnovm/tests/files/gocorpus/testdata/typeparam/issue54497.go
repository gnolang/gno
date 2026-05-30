// errorcheck -0 -m

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that inlining works with generic functions.

package testcase

type C interface{ ~uint | ~uint32 | ~uint64 }

func isAligned[T C](x, y T) bool { // ERROR "can inline isAligned\[uint\]" "can inline isAligned\[go\.shape\.uint\]" "inlining call to isAligned\[go\.shape\.uint\]"
	return x%y == 0
}

func foo(x uint) bool { // ERROR "can inline foo"
	return isAligned(x, 64) // ERROR "inlining call to isAligned\[go\.shape\.uint\]"
}

// GnoIncomplete: covered 0 of 3 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// KnownIssue:
// line 11: checker for ILLEGAL does not exist
