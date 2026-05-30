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

// GnoOutput:

// GoOutput:
// 1
// 2
// 3

// Divergence: TODO: <category>: explain why this divergence is acceptable