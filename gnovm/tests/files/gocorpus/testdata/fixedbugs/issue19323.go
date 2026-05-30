// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func g() {}

func f() {
	g()[:] // ERROR "g.* used as value|attempt to slice object that is not"
}

func g2() ([]byte, []byte) { return nil, nil }

func f2() {
	g2()[:] // ERROR "multiple-value g2.* in single-value context|attempt to slice object that is not|2\-valued g"
}

// GnoError:
// line 12: cannot slice variable of type ()
// line 18: cannot slice variable of type ([]uint8,[]uint8)

// GoTypeCheckError:
// line 12: g() (no value) used as value
// line 18: multiple-value g2() (value of type ([]byte, []byte)) in single-value context
