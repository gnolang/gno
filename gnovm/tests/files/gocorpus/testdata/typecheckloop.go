// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that constant definition loops are caught during
// typechecking and that the errors print correctly.

package main

const A = 1 + B // ERROR "constant definition loop\n.*A uses B\n.*B uses C\n.*C uses A|initialization cycle"
const B = C - 1 // ERROR "constant definition loop\n.*B uses C\n.*C uses B|initialization cycle"
const C = A + B + 1

// GnoError:
// line 14: initialization cycle: A refers to B refers to C refers to A

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 12, 13; over-strict: lines 14; on-chain path first error on line 12, not reported by accumulate mode
