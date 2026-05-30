// errorcheck -0 -m

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Make sure the go:noinline pragma makes it from a
// generic function to any of its stenciled instances.

package main

//go:noinline
func f[T any](x T) T {
	return x
}

func main() { // ERROR "can inline main"
	println(f(5))
}

// GnoIncomplete: covered 0 of 1 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// KnownIssue:
// line 13: 2: name T not defined in fileset with files [pragma.go]
