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
// 1
// 2
// 3

// GoOutput:
// 1
// 2
// 3

// Verdict: Clean run
