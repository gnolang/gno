// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type A[T any] interface {
	m()
}

type Z struct {
	a,b int
}

func (z *Z) m() {
}

func test[T any]() {
	var a A[T] = &Z{}
	f := a.m
	f()
}
func main() {
	test[string]()
}

// GnoOutput:

// GnoError:
// main/issue49049.go:21:6-19: name T not defined in fileset with files [issue49049.go]

// GoOutput:

// Unsupported: generics not supported in Gno
