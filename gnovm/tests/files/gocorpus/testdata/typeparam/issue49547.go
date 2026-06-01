// run

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

type foo int

func main() {
	want := "main.F[main.foo]"
	got := fmt.Sprintf("%T", F[foo]{})
	if got != want {
		fmt.Printf("want: %s, got: %s\n", want, got)
	}
}

type F[T any] struct {
}

// GnoOutput:

// GnoError:
// main/issue49547.go:15:27-33: unexpected index base type type{} (*gnolang.TypeType base *gnolang.TypeType)

// GoOutput:

// Unsupported: generics not supported in Gno
