// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that initialization loops are caught
// and that the errors print correctly.

package main

var (
	x int = a
	a int = b // ERROR "a refers to b\n.*b refers to c\n.*c refers to a|initialization loop"
	b int = c
	c int = a
)

// GnoError:
// line 13: invalid recursive value: x -> a -> b -> c -> a
// line 14: invalid recursive value: a -> b -> c -> a
// line 15: decl node already predefined! var a<!VPInvalid(0)> int<VPInvalid(0)> = b<VPInvalid(0)>
// line 16: decl node already predefined! var a<!VPInvalid(0)> int<VPInvalid(0)> = b<VPInvalid(0)>

// Verdict: KnownIssue errorcheck: #6344; over-strict: lines 13; internal errors: lines 15, 16
