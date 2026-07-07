// run

// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	m := "0"
	for _, c := range "321" {
		m = max(string(c), m)
		println(m)
	}
}

// KnownDivergence:
// Go1.17 pin.

// GnoOutput:

// GnoError:
// main/issue64565.go:12:7-10: name max not declared

// GoOutput:
// 3
// 3
// 3
