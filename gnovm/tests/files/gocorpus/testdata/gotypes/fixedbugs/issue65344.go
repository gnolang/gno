// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T1 C /* ERROR "C is not a type" */

// TODO(gri) try to avoid this follow-on error
const C = T1(0 /* ERROR "cannot convert 0 (untyped int constant) to type T1" */)

type T2 V /* ERROR "V is not a type" */

var V T2

func _() {
	// don't produce errors here
	_ = C + V
}

// GnoError:
// line 10: cannot convert 0 to type gno.land/p/filetest/p.T1: gno.land/p/filetest/p.T1 refers to itself

// GnoWordingDiffers:
// line 10: gc expects: cannot convert 0 (untyped int constant) to type T1; gno: cannot convert 0 to type gno.land/p/filetest/p.T1: gno.land/p/filetest/p.T1 refers to itself

// Verdict: KnownIssue errorcheck: #6344; uncaught markers: lines 7, 12
