// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type myint int
func (x myint) foo() int {return int(x)}

type myfloat float64
func (x myfloat) foo() float64 {return float64(x) }

func f[T any](i interface{}) {
	switch x := i.(type) {
	case interface { foo() T }:
		println("fooer", x.foo())
	default:
		println("other")
	}
}
func main() {
	f[int](myint(6))
	f[int](myfloat(7))
	f[float64](myint(8))
	f[float64](myfloat(9))
}

// GnoOutput:

// GnoError:
// main/typeswitch5.go:17:25-26: name T not declared

// GoOutput:
// fooer 6
// other
// other
// fooer +9.000000e+000

// Unsupported: generics not supported in Gno
