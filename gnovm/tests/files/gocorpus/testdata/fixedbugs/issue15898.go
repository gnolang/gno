// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func f(e interface{}) {
	switch e.(type) {
	case nil, nil: // ERROR "multiple nil cases in type switch|duplicate type in switch|duplicate case nil in type switch"
	}

	switch e.(type) {
	case nil:
	case nil: // ERROR "multiple nil cases in type switch|duplicate type in switch|duplicate case nil in type switch"
	}
}

// GnoError:
// line 10: 3: duplicate type nil in type switch
// line 11: expected '}', found 'case' (and 2 more errors)
// line 14: expected declaration, found 'switch' (and 1 more errors)
// line 15: expected declaration, found 'case'
// line 16: expected declaration, found 'case'
// line 17: expected declaration, found '}'
// line 18: expected declaration, found '}'

// GoTypeCheckError:
// line 11: duplicate case nil in type switch
// line 16: duplicate case nil in type switch

// GnoOverStrictError:
// line 10: 3: duplicate type nil in type switch
// line 14: expected declaration, found 'switch' (and 1 more errors)
// line 15: expected declaration, found 'case'
// line 17: expected declaration, found '}'
// line 18: expected declaration, found '}'
