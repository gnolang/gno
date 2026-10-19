// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 4470: parens are not allowed around .(type) "expressions"

package main

func main() {
	var i interface{}
	switch (i.(type)) { // ERROR "outside type switch"
	default:
	}
	_ = i
}

// GnoError:
// line 13: should not happen

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 13; internal errors: lines 13
