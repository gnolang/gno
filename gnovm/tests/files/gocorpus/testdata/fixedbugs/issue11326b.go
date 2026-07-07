// run

// Does not work with gccgo, which uses a smaller (but still permitted)
// exponent size.
//go:build !gccgo

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Tests for golang.org/issue/11326.

func main() {
	{
		const n = 1e646456992
		const d = 1e646456991
		x := n / d
		if x != 10.0 {
			println("incorrect value:", x)
		}
	}
	{
		const n = 1e64645699
		const d = 1e64645698
		x := n / d
		if x != 10.0 {
			println("incorrect value:", x)
		}
	}
	{
		const n = 1e6464569
		const d = 1e6464568
		x := n / d
		if x != 10.0 {
			println("incorrect value:", x)
		}
	}
	{
		const n = 1e646456
		const d = 1e646455
		x := n / d
		if x != 10.0 {
			println("incorrect value:", x)
		}
	}
}

// GnoOutput:

// GnoError:
// main/issue11326b.go:17:13-24: invalid decimal constant: 1e646456992

// GoOutput:

// KnownDivergence:
// Untyped float constants are backed by apd.Decimal, whose exponent is
// capped at ±100000, so 1e646456992 is unrepresentable (op_eval.go FLOAT).
// Spec only requires 16-bit constant exponents (1e32767 works on master),
// so this is a permitted implementation limit — same class as gccgo, which
// this test also skips.
