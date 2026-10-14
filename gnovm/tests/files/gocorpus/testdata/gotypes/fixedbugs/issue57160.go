// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _(x *int) {
	_ = 0 < x // ERROR "invalid operation"
	_ = x < 0 // ERROR "invalid operation"
}

// GnoError:
// line 8: operator < not defined on: PointerKind
// line 9: operator < not defined on: PointerKind

// GoTypeCheckError:
// line 8
// line 9

// GnoWordingDiffers:
// line 8: gc expects: invalid operation; gno: operator < not defined on: PointerKind
// line 9: gc expects: invalid operation; gno: operator < not defined on: PointerKind

// Verdict: Clean errorcheck
