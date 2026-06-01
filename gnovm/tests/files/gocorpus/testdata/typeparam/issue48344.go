// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type G[T any] interface {
	g()
}

type Foo[T any] struct {
}

func (foo *Foo[T]) g() {

}

func f[T any]() {
	v := []G[T]{}
	v = append(v, &Foo[T]{})
}
func main() {
	f[int]()
}

// GnoOutput:

// GnoError:
// main/issue48344.go:16:1-18:2: name T not defined in fileset with files [issue48344.go]

// GoOutput:

// Unsupported: generics not supported in Gno
