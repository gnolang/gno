// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type bar struct {
	x int
}

func main() {
	var foo bar
	_ = &foo{} // ERROR "is not a type|expected .;."
} // GCCGO_ERROR "expected declaration"

// GnoError:
// line 15: foo<VPBlock(1,0)> is not a type

// GoTypeCheckError:
// line 15: foo is not a type
