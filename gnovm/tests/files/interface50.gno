// run
// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Check that taking a method value from a nil interface (selector without
// call) produces a proper panic, not just method invocation.
package main

type I interface {
	M()
}

func main() {
	defer func() {
		err := recover()
		if err == nil {
			panic("panic expected")
		}
		println(err)
	}()
	var i I
	f := i.M
	_ = f
}

// Output:
// runtime error: method selector on nil interface
