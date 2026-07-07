// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main
const ( F = 1 )
func fn(i int) int {
	if i == F() {		// ERROR "func"
		return 0
	}
	return 1
}

// GnoError:
// line 10: unexpected func type <untyped> bigint (gnolang.PrimitiveType)
// line 13: expected declaration, found 'return'
// line 14: expected declaration, found '}'

// GoTypeCheckError:
// line 10: invalid operation: cannot call F (untyped int constant 1): untyped int is not a function

// GnoOverStrictError:
// line 13: expected declaration, found 'return'
// line 14: expected declaration, found '}'
