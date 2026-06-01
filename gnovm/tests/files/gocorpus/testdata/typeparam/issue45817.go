// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
)

type s[T any] struct {
	a T
}

func (x s[T]) f() T {
	return x.a
}
func main() {
	x := s[int]{a: 7}
	f := x.f
	if got, want := f(), 7; got != want {
		panic(fmt.Sprintf("got %d, want %d", got, want))
	}
}

// GnoOutput:

// GnoError:
// main/issue45817.go:13:6-15:2: name T not defined in fileset with files [issue45817.go]

// GoOutput:

// Unsupported: generics not supported in Gno
