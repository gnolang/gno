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
// line 10: expected declaration, found 0
// line 11: expected declaration, found 0
// line 12: expected declaration, found '}'
// line 14: 2: duplicate key (0 int) in map literal
// line 15: expected declaration, found 'interface'
// line 16: expected declaration, found 'interface'
// line 17: expected declaration, found '}'

// GoTypeCheckError:
// line 11: duplicate key 0 in map literal
// line 22: duplicate case 0 (constant of type int) in expression switch

// GnoOverStrictError:
// line 9: 2: duplicate key (0 int) in map literal
// line 10: expected declaration, found 0
// line 12: expected declaration, found '}'
// line 14: 2: duplicate key (0 int) in map literal
// line 15: expected declaration, found 'interface'
// line 16: expected declaration, found 'interface'
// line 17: expected declaration, found '}'
