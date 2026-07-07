// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type t struct {
	x int // GCCGO_ERROR "duplicate field name .x."
	x int // GC_ERROR "duplicate field x|x redeclared"
}

func f(t *t) int {
	return t.x
}

// Unsupported: only gc-specific (GC_ERROR) markers; not part of Gno's contract
