// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _() {
	type Bool bool
	for range func /* ERROR "yield func returns user-defined boolean, not bool" */ (func() Bool) {} {
	}
	for range func /* ERROR "yield func returns user-defined boolean, not bool" */ (func(int) Bool) {} {
	}
	for range func /* ERROR "yield func returns user-defined boolean, not bool" */ (func(int, string) Bool) {} {
	}
}

// GnoError:
// line 9: range iteration requires map, string, array, slice, or pointer to array; got FuncKind
// line 11: range iteration requires map, string, array, slice, or pointer to array; got FuncKind
// line 13: range iteration requires map, string, array, slice, or pointer to array; got FuncKind

// GnoWordingDiffers:
// line 9: gc expects: yield func returns user-defined boolean, not bool; gno: range iteration requires map, string, array, slice, or pointer to array; got FuncKind
// line 11: gc expects: yield func returns user-defined boolean, not bool; gno: range iteration requires map, string, array, slice, or pointer to array; got FuncKind
// line 13: gc expects: yield func returns user-defined boolean, not bool; gno: range iteration requires map, string, array, slice, or pointer to array; got FuncKind

// Verdict: Clean errorcheck
