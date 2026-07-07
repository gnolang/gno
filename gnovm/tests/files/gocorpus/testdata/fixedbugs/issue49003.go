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
// line 9: 2: [function "f" does not terminate]
// line 10: expected declaration, found 'for'
// line 11: expected declaration, found '}'
// line 12: expected declaration, found '}'

// GoTypeCheckError:
// line 12: missing return

// GnoOverStrictError:
// line 9: 2: [function "f" does not terminate]
// line 10: expected declaration, found 'for'
// line 11: expected declaration, found '}'
