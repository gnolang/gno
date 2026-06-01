// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type SomeInterface interface {
	Whatever()
}

func X[T any]() T {
	var m T

	// for this example, this block should never run
	if _, ok := any(m).(SomeInterface); ok {
		var dst SomeInterface
		_, _ = dst.(T)
		return dst.(T)
	}

	return m
}

type holder struct{}

func main() {
	X[holder]()
}

// GnoOutput:

// GnoError:
// main/issue52228.go:13:1-24:2: name T not defined in fileset with files [issue52228.go]

// GoOutput:

// Unsupported: generics not supported in Gno
