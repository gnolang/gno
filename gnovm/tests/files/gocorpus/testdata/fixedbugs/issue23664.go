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
// line 12: expected ';', found true (and 2 more errors)
// line 15: expected declaration, found 'switch'
