// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type I interface {
	[]byte
}

func F[T I]() {
	var t T
	explodes(t)
}

func explodes(b []byte) {}

func main() {

}

// GnoOutput:

// GnoError:
// main/issue51236.go:9:8-11:2: unexpected field type []uint8

// GoOutput:

// Unsupported: generics not supported in Gno
