// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f(int, int) {
    switch x {
    case 1:
        f(1, g()   // ERROR "expecting \)|possibly missing comma or \)"
    case 2:
        f()
    case 3:
        f(1, g()   // ERROR "expecting \)|possibly missing comma or \)"
    }
}

// GnoError:
// line 12: missing ',' before newline in argument list
// line 13: expected operand, found 'case'
// line 18: missing ',' in argument list

// ParseBlocked:
// line 16: gc expects: expecting \)|possibly missing comma or \)

// GnoWordingDiffers:
// line 12: gc expects: expecting \)|possibly missing comma or \); gno: missing ',' before newline in argument list

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 16; over-strict: lines 13, 18
