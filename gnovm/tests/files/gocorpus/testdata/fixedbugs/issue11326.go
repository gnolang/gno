// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tests for golang.org/issue/11326.

package main

func main() {
	// The gc compiler implementation uses the minimally required 32bit
	// binary exponent, so these constants cannot be represented anymore
	// internally. However, the language spec does not preclude other
	// implementations from handling these. Don't check the error.
	// var _ = 1e2147483647 // "constant too large"
	// var _ = 1e646456993  // "constant too large"

	// Any implementation must be able to handle these constants at
	// compile time (even though they cannot be assigned to a float64).
	var _ = 1e646456992  // ERROR "1e\+646456992 overflows float64|floating-point constant overflow|exponent too large|overflows float64|overflows"
	var _ = 1e64645699   // ERROR "1e\+64645699 overflows float64|floating-point constant overflow|exponent too large|overflows float64|overflows"
	var _ = 1e6464569    // ERROR "1e\+6464569 overflows float64|floating-point constant overflow|exponent too large|overflows float64|overflows"
	var _ = 1e646456     // ERROR "1e\+646456 overflows float64|floating-point constant overflow|exponent too large|overflows float64|overflows"
	var _ = 1e64645      // ERROR "1e\+64645 overflows float64|floating-point constant overflow|exponent too large|overflows float64|overflows"
	var _ = 1e6464       // ERROR "1e\+6464 overflows float64|floating-point constant overflow|overflows float64|overflows"
	var _ = 1e646        // ERROR "1e\+646 overflows float64|floating-point constant overflow|overflows float64|overflows"
	var _ = 1e309        // ERROR "1e\+309 overflows float64|floating-point constant overflow|overflows float64|overflows"

	var _ = 1e308
}

// GnoError:
// line 21: invalid decimal constant: 1e646456992
// line 22: invalid decimal constant: 1e64645699
// line 23: invalid decimal constant: 1e6464569
// line 24: invalid decimal constant: 1e646456

// GoTypeCheckError:
// line 21: cannot use 1e646456992 (untyped float constant 1e+646456992) as float64 value in variable declaration (overflows)
// line 22: cannot use 1e64645699 (untyped float constant 1e+64645699) as float64 value in variable declaration (overflows)
// line 23: cannot use 1e6464569 (untyped float constant 1e+6464569) as float64 value in variable declaration (overflows)
// line 24: cannot use 1e646456 (untyped float constant 1e+646456) as float64 value in variable declaration (overflows)
// line 25: cannot use 1e64645 (untyped float constant 1e+64645) as float64 value in variable declaration (overflows)
// line 26: cannot use 1e6464 (untyped float constant 1e+6464) as float64 value in variable declaration (overflows)
// line 27: cannot use 1e646 (untyped float constant 1e+646) as float64 value in variable declaration (overflows)
// line 28: cannot use 1e309 (untyped float constant 1e+309) as float64 value in variable declaration (overflows)
