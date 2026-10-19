// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 4359: wrong handling of broken struct fields
// causes "internal compiler error: lookdot badwidth".

package main

type T struct {
	x T1 // ERROR "undefined"
}

func f() {
	var t *T
	_ = t.x
}

// GnoError:
// line 13: name not defined: T1

// Verdict: KnownIssue errorcheck: #6344 on-chain path first error on line 12, not reported by accumulate mode
