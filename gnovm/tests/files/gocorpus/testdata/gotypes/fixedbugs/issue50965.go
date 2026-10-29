// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _(x int, c string) {
	switch x {
	case c /* ERROR "invalid case c in switch on x (mismatched types string and int)" */ :
	}
}

func _(x, c []int) {
	switch x {
	case c /* ERROR "invalid case c in switch on x (slice can only be compared to nil)" */ :
	}
}

// GnoError:
// line 9: cannot use string as int
// line 15: []int can only be compared to nil

// GnoWordingDiffers:
// line 9: gc expects: invalid case c in switch on x (mismatched types string and int); gno: cannot use string as int
// line 15: gc expects: invalid case c in switch on x (slice can only be compared to nil); gno: []int can only be compared to nil

// Verdict: Clean errorcheck
