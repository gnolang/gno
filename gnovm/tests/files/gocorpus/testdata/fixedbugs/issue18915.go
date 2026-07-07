// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Make sure error message for invalid conditions
// or tags are consistent with earlier Go versions.

package p

func _() {
	if a := 10 { // ERROR "cannot use a := 10 as value|expected .*;|declared and not used"
	}

	for b := 10 { // ERROR "cannot use b := 10 as value|parse error|declared and not used"
	}

	switch c := 10 { // ERROR "cannot use c := 10 as value|expected .*;|declared and not used"
	}
}

// GnoError:
// line 13: expected boolean expression, found assignment (missing parentheses around composite literal?) (and 2 more errors)
// line 16: expected declaration, found 'for'
// line 17: expected declaration, found '}'
// line 19: expected declaration, found 'switch'
// line 20: expected declaration, found '}'
// line 21: expected declaration, found '}'

// GoTypeCheckError:
// line 13: expected boolean expression, found assignment (missing parentheses around composite literal?) (and 2 more errors)
// line 16: expected declaration, found 'for'
// line 19: expected declaration, found 'switch'

// GnoOverStrictError:
// line 17: expected declaration, found '}'
// line 20: expected declaration, found '}'
// line 21: expected declaration, found '}'
