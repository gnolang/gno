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
// line 10: unexpected index base kind for type type{}
// line 11: unexpected index base kind for type type{}
// line 12: unexpected index base kind for type type{}
// line 13: unexpected index base kind for type type{}
// line 18: expected type, found 1

// GnoWordingDiffers:
// line 11: gc expects: cannot use.*without instantiation; gno: unexpected index base kind for type type{}
// line 12: gc expects: invalid operation: bool[int] (bool is not a generic type); gno: unexpected index base kind for type type{}
// line 13: gc expects: invalid operation: bool[G] (bool is not a generic type); gno: unexpected index base kind for type type{}

// Verdict: KnownIssue errorcheck: #6344 over-strict: lines 10
