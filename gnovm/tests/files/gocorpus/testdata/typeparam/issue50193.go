// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
)

type Complex interface {
	~complex64 | ~complex128
}

func zero[T Complex]() T {
	return T(0)
}
func pi[T Complex]() T {
	return T(3.14)
}
func sqrtN1[T Complex]() T {
	return T(-1i)
}

func main() {
	fmt.Println(zero[complex128]())
	fmt.Println(pi[complex128]())
	fmt.Println(sqrtN1[complex128]())
	fmt.Println(zero[complex64]())
	fmt.Println(pi[complex64]())
	fmt.Println(sqrtN1[complex64]())
}

// GnoOutput:

// GnoError:
// main/issue50193.go:13:6-15:2: name complex64 not defined in fileset with files [issue50193.go]

// GoOutput:
// (0+0i)
// (3.14+0i)
// (0-1i)
// (0+0i)
// (3.14+0i)
// (0-1i)

// Unsupported: generics not supported in Gno
