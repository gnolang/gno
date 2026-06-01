// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type myifacer[T any] interface{ do(T) error }

type stuff[T any] struct{}

func (s stuff[T]) run() interface{} {
	var i myifacer[T]
	return i
}

func main() {
	stuff[int]{}.run()
}

// GnoOutput:

// GnoError:
// main/issue47925.go:9:6-46: name T not defined in fileset with files [issue47925.go]

// GoOutput:

// Unsupported: generics not supported in Gno
