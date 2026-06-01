// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type Constraint[T any] interface {
	~func() T
}

func Foo[T Constraint[T]]() T {
	var t T

	t = func() T {
		return t
	}
	return t
}

func main() {
	type Bar func() Bar
	Foo[Bar]()
}

// GnoOutput:

// GnoError:
// main/issue48137.go:9:6-11:2: name T not defined in fileset with files [issue48137.go]

// GoOutput:

// Unsupported: generics not supported in Gno
