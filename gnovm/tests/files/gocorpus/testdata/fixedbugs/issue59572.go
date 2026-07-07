// run

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func foo() {
	println("foo")
}

func main() {
	fn := foo
	for _, fn = range list {
		fn()
	}
}

var list = []func(){
	func() {
		println("1")
	},
	func() {
		println("2")
	},
	func() {
		println("3")
	},
}


// Fixed: master PR #5764 (98f4db57c); verified 1/2/3 output, broken at parent; re-golden after rebase.

// GnoOutput:

// GnoError:
// runtime error: invalid memory address or nil pointer dereference

// GoOutput:
// 1
// 2
// 3

// KnownIssue:
// for _, fn = range (blank key, assignment to an outer var) crashed
// preprocess with a nil-type deref on the blank operand — the nil-iff-blank
// range-operand handling. Same root cause as fixedbugs/bug406.go.
