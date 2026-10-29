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
// line 14: initialization cycle: a refers to b refers to c refers to a

// GnoWordingDiffers:
// line 14: gc expects: a refers to b
// .*b refers to c
// .*c refers to a|initialization loop; gno: initialization cycle: a refers to b refers to c refers to a

// Verdict: Clean errorcheck
