// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f(s string) int {
	for range s {
	}
} // ERROR "missing return"

// GnoError:
// line 9: [function "f" does not terminate]

// GnoWordingDiffers:
// line 12: gc expects: missing return; gno: 2: [function "f" does not terminate]

// Verdict: Clean errorcheck
