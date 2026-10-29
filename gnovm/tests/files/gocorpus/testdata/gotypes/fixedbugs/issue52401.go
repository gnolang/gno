// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _() {
	const x = 0
	x /* ERROR "cannot assign to x" */ += 1
	x /* ERROR "cannot assign to x" */ ++
}

// GnoError:
// line 9: cannot assign to const x
// line 10: cannot assign to x

// GnoWordingDiffers:
// line 9: gc expects: cannot assign to x; gno: cannot assign to const x

// Verdict: Clean errorcheck
