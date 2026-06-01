// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	check[string]()
}

func check[T any]() {
	var result setter[T]
	switch result.(type) {
	case fooA[T]:
	case fooB[T]:
	}
}

type setter[T any] interface {
	Set(T)
}

type fooA[T any] struct{}

func (fooA[T]) Set(T) {}

type fooB[T any] struct{}

func (fooB[T]) Set(T) {}

// GnoOutput:

// GnoError:
// main/issue48838.go:21:6-23:2: name T not defined in fileset with files [issue48838.go]

// GoOutput:

// Unsupported: generics not supported in Gno
