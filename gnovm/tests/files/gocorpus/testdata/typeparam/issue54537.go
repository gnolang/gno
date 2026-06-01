// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	_ = F[bool]

	var x string
	_ = G(x == "foo")
}

func F[T ~bool](x string) {
	var _ T = x == "foo"
}

func G[T any](t T) *T {
	return &t
}

// GnoOutput:

// GnoError:
// main/issue54537.go:20:1-22:2: name T not defined in fileset with files [issue54537.go]

// GoOutput:

// Unsupported: generics not supported in Gno
