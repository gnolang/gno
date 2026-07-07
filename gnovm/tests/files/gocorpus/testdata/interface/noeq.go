// run

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test run-time error detection for interface values containing types
// that cannot be compared for equality.

package main

func main() {
	cmp(1)

	var (
		m map[int]int
		s struct{ x []int }
		f func()
	)
	noCmp(m)
	noCmp(s)
	noCmp(f)
}

func cmp(x interface{}) bool {
	return x == x
}

func noCmp(x interface{}) {
	shouldPanic(func() { cmp(x) })
}

func shouldPanic(f func()) {
	defer func() {
		if recover() == nil {
			panic("function should panic")
		}
	}()
	f()
}


// Fixed: master PR #5713 (5d889b083); verified clean, broken at parent; re-golden after rebase.

// GnoOutput:

// GnoError:
// function should panic

// GoOutput:

// KnownIssue:
// Comparing interface values with uncomparable dynamic types (map, func,
// struct containing a slice) returned false instead of panicking at
// runtime. Same root cause as fixedbugs/issue8606.go.
