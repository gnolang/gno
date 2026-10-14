// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _() {
	var x int
	var f func() []int
	_ = f /* ERROR "cannot index f" */ [x]
}

// GnoError:
// line 10: unexpected index base kind for type func() []int

// GoTypeCheckError:
// line 10

// GnoWordingDiffers:
// line 10: gc expects: cannot index f; gno: unexpected index base kind for type func() []int

// Verdict: Clean errorcheck
