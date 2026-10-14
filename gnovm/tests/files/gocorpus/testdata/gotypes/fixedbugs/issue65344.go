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

// GoTypeCheckError:
// line 7
// line 10
// line 12

// GoTypesOnly:
// line 7: gc expects: C is not a type
// line 10: gc expects: cannot convert 0 (untyped int constant) to type T1
// line 12: gc expects: V is not a type

// Verdict: Clean errorcheck
