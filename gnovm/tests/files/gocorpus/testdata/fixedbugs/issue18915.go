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
// line 13: expected boolean expression, found assignment (missing parentheses around composite literal?)

// GoTypeCheckError:
// line 13

// UncaughtError:
// line 16: uncaught; gc expects: cannot use b := 10 as value|parse error|declared and not used
// line 19: uncaught; gc expects: cannot use c := 10 as value|expected .*;|declared and not used

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 2 uncaught marker(s)
