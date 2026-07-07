// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f(x int) {
	switch x {
	case 0:
		fallthrough
		; // ok
	case 1:
		fallthrough // ERROR "fallthrough statement out of place"
		{}
	case 2:
		fallthrough // ERROR "cannot fallthrough"
	}
}

// GnoError:
// line 12: fallthrough statement out of place
// line 15: fallthrough statement out of place
// line 18: cannot fallthrough final case in switch

// GoTypeCheckError:
// line 15: fallthrough statement out of place
// line 18: cannot fallthrough final case in switch

// GnoOverStrictError:
// line 12: fallthrough statement out of place
