// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify error messages for incorrect if/switch headers.

package p

func f() {
	if f() true { // ERROR "unexpected name true, expected {"
	}

	switch f() true { // ERROR "unexpected name true, expected {"
	}
}

// GnoError:
// line 12: expected ';', found true

// GoTypeCheckError:
// line 12

// ParseBlocked:
// line 15: gc expects: unexpected name true, expected {

// GnoWordingDiffers:
// line 12: gc expects: unexpected name true, expected {; gno: expected ';', found true

// Verdict: Clean errorcheck
