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
	case int32, int16:
		println("int32/int16")
	case struct{ a, b T }:
		println("struct{T,T}")
	default:
		println("other")
	}
}
func main() {
	f[float64](float64(6))
	f[float64](int(7))
	f[float64](int32(8))
	f[float64](struct{ a, b float64 }{a: 1, b: 2})
	f[float64](int8(9))
	f[int32](int32(7))
	f[int](int32(7))
	f[any](int(10))
	f[interface{ M() }](int(11))
}

// GnoOutput:

// GnoError:
// main/typeswitch1.go:11:7-8: name T not declared

// GoOutput:
// T
// int
// int32/int16
// struct{T,T}
// other
// T
// int32/int16
// T
// int

// Unsupported: generics not supported in Gno
