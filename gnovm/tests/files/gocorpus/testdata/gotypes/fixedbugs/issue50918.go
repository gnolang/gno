// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type thing1 struct {
	things []string
}

type thing2 struct {
	things []thing1
}

func _() {
	var a1, b1 thing1
	_ = a1 /* ERROR "struct containing []string cannot be compared" */ == b1

	var a2, b2 thing2
	_ = a2 /* ERROR "struct containing []thing1 cannot be compared" */ == b2
}

// GnoError:
// line 17: gno.land/p/filetest/p.thing1 is not comparable
// line 20: gno.land/p/filetest/p.thing2 is not comparable

// GoTypeCheckError:
// line 17
// line 20

// Clean: errorcheck
