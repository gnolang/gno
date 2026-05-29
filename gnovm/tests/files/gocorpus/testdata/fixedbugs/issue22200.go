// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f1(x *[1<<30 - 1e6]byte) byte {
	for _, b := range *x {
		return b
	}
	return 0
}
func f2(x *[1<<30 + 1e6]byte) byte { // GC_ERROR "stack frame too large"
	for _, b := range *x {
		return b
	}
	return 0
}

// Unsupported: Gno doesn't perform gc's "stack frame too large" analysis; it accepts this file (gc rejects it), so there's no error to pin.
