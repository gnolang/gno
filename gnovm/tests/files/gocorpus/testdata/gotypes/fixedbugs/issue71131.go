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
// line 9: 3: range iteration requires map, string, array, slice, or pointer to array; got FuncKind
// line 11: 3: range iteration requires map, string, array, slice, or pointer to array; got FuncKind
// line 13: 3: range iteration requires map, string, array, slice, or pointer to array; got FuncKind

// GoTypeCheckError:
// line 9
// line 11
// line 13

// Clean: errorcheck
