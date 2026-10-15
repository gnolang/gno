// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 4458: gc accepts invalid method expressions
// like (**T).Method.

package main

type T struct{}

func (T) foo() {}

func main() {
	av := T{}
	pav := &av
	(**T).foo(&pav) // ERROR "no method .*foo|requires named type or pointer to named|undefined"
}

// GnoError:
// line 19: expected block type value path but got VPInvalid:
// --- preprocess stack ---
// stack 2: func main/issue4458.go:16:1-20:2
// ------------------------

// Verdict: Clean errorcheck
