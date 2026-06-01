// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type C[T any] struct {
}

func (c *C[T]) reset() {
}

func New[T any]() {
	c := &C[T]{}
	i = c.reset
	z(c.reset)
}

var i interface{}

func z(interface{}) {
}

func main() {
	New[int]()
}

// GnoOutput:

// GnoError:
// main/issue47775b.go:12:1-13:2: name T not defined in fileset with files [issue47775b.go]

// GoOutput:

// Unsupported: generics not supported in Gno
