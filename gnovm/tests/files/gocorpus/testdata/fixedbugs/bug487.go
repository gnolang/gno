// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The gccgo compiler did not reliably report mismatches between the
// number of function results and the number of expected results.

package p

func G() (int, int, int) {
	return 0, 0, 0
}

func F() {
	a, b := G()	// ERROR "mismatch|cannot initialize"
	a, b = G()	// ERROR "mismatch|cannot assign"
	_, _ = a, b
}

func H() (int, int) {
	return G()	// ERROR "too many|mismatch|wrong number"
}

// GnoError:
// line 17: assignment mismatch: 2 variables but G<VPBlock(3,0)> returns 3 values
// line 18: assignment mismatch: 2 variables but G<VPBlock(3,0)> returns 3 values
// line 23: expected 2 return values; got 3

// GnoWordingDiffers:
// line 23: gc expects: too many|mismatch|wrong number; gno: expected 2 return values; got 3

// Verdict: Clean errorcheck
