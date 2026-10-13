// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type G[P any] int

type (
	_ G[int]
	_ G[G /* ERRORx `cannot use.*without instantiation` */]
	_ bool /* ERROR "invalid operation: bool[int] (bool is not a generic type)" */ [int]
	_ bool /* ERROR "invalid operation: bool[G] (bool is not a generic type)" */[G]
)

// The example from the issue.
func _() {
	_ = &([10]bool /* ERRORx `invalid operation.*bool is not a generic type` */ [1 /* ERROR "expected type" */ ]{})
}

// GnoError:
// line 18: expected type, found 1

// GoTypeCheckError:
// line 18

// UncaughtError:
// line 11: uncaught; gc expects: cannot use.*without instantiation
// line 12: uncaught; gc expects: invalid operation: bool[int] (bool is not a generic type)
// line 13: uncaught; gc expects: invalid operation: bool[G] (bool is not a generic type)

// KnownIssue:
// errorcheck: 0 over-strict line(s), 3 uncaught marker(s)
