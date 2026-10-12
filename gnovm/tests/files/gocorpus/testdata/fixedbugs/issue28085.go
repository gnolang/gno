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

// GnoError:
// line 9: 2: duplicate key (0 int) in map literal
// line 14: 2: duplicate key (0 int) in map literal

// GoTypeCheckError:
// line 11
// line 21
// line 22

// GoTypesOnly:
// line 22: gc expects: duplicate

// GnoOverStrictError:
// line 14: 2: duplicate key (0 int) in map literal

// Verdict: KnownIssue errorcheck: 1 over-strict line(s), 0 uncaught marker(s)
