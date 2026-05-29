// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f() bool {
	if false {
	} else {
		return true
	}
} // ERROR "missing return( at end of function)?"

// GnoError:
// line 9: 2: [function "f" does not terminate]

// GoTypeCheckError:
// line 14: missing return
