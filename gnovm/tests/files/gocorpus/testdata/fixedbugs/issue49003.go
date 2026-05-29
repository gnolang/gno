// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f(s string) int {
	for range s {
	}
} // ERROR "missing return"

// GnoIncomplete: covered 0 of 1 markers; Gno bailed before the rest — a runnable variant is needed to exercise them
// GnoError:
// line 9: 2: [function "f" does not terminate]
