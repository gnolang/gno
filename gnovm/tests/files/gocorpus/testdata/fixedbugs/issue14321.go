// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that error message reports _ambiguous_ method.

package p

type A struct{
	H int
}

func (A) F() {}
func (A) G() {}

type B struct{
	G int
	H int
}

func (B) F() {}

type C struct {
	A
	B
}

var _ = C.F // ERROR "ambiguous"
var _ = C.G // ERROR "ambiguous"
var _ = C.H // ERROR "ambiguous"
var _ = C.I // ERROR "no method .*I.*|C.I undefined"

// GnoError:
// line 30: ambiguous selector C.F
// line 31: ambiguous selector C.G
// line 32: ambiguous selector C.H
// line 33: C.I undefined (type C has no field or method I)
