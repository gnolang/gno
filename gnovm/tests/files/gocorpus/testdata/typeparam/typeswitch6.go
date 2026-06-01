// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f[T any](i interface{}) {
	switch i.(type) {
	case T:
		println("T")
	case int:
		println("int")
	default:
		println("other")
	}
}

type myint int
func (myint) foo() {
}

func main() {
	f[interface{}](nil)
	f[interface{}](6)
	f[interface{foo()}](nil)
	f[interface{foo()}](7)
	f[interface{foo()}](myint(8))
}

// GnoOutput:

// GnoError:
// main/typeswitch6.go:11:7-8: name T not declared

// GoOutput:
// other
// T
// other
// int
// T

// Unsupported: generics not supported in Gno
