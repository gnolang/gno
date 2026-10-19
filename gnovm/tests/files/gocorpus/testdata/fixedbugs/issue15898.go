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
// line 10: duplicate type nil in type switch
// line 14: duplicate type nil in type switch

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 11, 16; over-strict: lines 10, 14
