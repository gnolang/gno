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
// line 10: invalid operation: cannot index f (variable of type func() []int)

// Verdict: Clean errorcheck
