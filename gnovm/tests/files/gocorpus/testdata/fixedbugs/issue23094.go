// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that the array is reported in correct notation.

package p

var a [len(a)]int // ERROR "\[len\(a\)\]int|initialization cycle"

// GnoError:
// line 11: invalid recursive type: a -> a

// GnoWordingDiffers:
// line 11: gc expects: \[len\(a\)\]int|initialization cycle; gno: invalid recursive type: a -> a

// Verdict: Clean errorcheck
