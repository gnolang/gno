// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var _ = map[interface{}]int{
	0: 0,
	0: 0, // ERROR "duplicate"
}

var _ = map[interface{}]int{
	interface{}(0): 0,
	interface{}(0): 0, // ok
}

func _() {
	switch interface{}(0) {
	case 0:
	case 0: // ERROR "duplicate"
	}

	switch interface{}(0) {
	case interface{}(0):
	case interface{}(0): // ok
	}
}

// GnoOverStrictError:
// line 9: 2: duplicate key (0 int) in map literal

// GoTypeCheckError:
// line 11: duplicate key 0 in map literal
// line 22: duplicate case 0 (constant of type int) in expression switch

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)
