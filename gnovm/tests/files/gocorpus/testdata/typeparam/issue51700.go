// run

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f[B any](b B) {
	if b1, ok := any(b).(interface{ m1() }); ok {
		panic(1)
		_ = b1.(B)
	}
	if b2, ok := any(b).(interface{ m2() }); ok {
		panic(2)
		_ = b2.(B)
	}
}

type S struct{}

func (S) m3() {}

func main() {
	f(S{})
}

// GnoOutput:

// GnoError:
// main/issue51700.go:9:1-18:2: name B not defined in fileset with files [issue51700.go]

// GoOutput:

// Unsupported: generics not supported in Gno
