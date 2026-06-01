// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func Do[T any](do func() (T, string)) {
	_ = func() (T, string) {
		return do()
	}
}

func main() {
	Do[int](func() (int, string) {
		return 3, "3"
	})
}

// GnoOutput:

// GnoError:
// main/issue47514b.go:9:1-13:2: name T not defined in fileset with files [issue47514b.go]

// GoOutput:

// Unsupported: generics not supported in Gno
