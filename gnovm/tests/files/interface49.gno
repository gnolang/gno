// run
// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Check that calling a method on a nil interface causes a proper panic.
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
	i.M()
}

// Output:
// runtime error: method call on nil interface
