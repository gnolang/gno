// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f (x,		// GCCGO_ERROR "previous"
	x int) {	// ERROR "duplicate argument|redefinition|redeclared"
}

// GnoError:
// line 10: x redeclared in this block
// 	previous declaration at bug040.go:9:9

// GoTypeCheckError:
// line 10: x redeclared in this block
