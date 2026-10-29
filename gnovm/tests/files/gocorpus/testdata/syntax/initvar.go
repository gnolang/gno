// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	if var x = 0; x < 10 {}    // ERROR "var declaration not allowed in if initializer"

	switch var x = 0; x {}     // ERROR "var declaration not allowed in switch initializer"

	for var x = 0; x < 10; {}  // ERROR "var declaration not allowed in for initializer"
}

// GnoError:
// line 10: var declaration not allowed in if initializer
// line 12: expected '{', found 'var'
// line 14: expected '{', found 'var'
// line 15: expected ';', found 'EOF'

// Verdict: KnownIssue errorcheck: #6344 over-strict: lines 15
