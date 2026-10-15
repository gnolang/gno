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
// line 30: C.F undefined (type gno.land/p/filetest/p.C has no method F)
// line 31: C.G undefined (type gno.land/p/filetest/p.C has no method G)
// line 32: C.H undefined (type gno.land/p/filetest/p.C has no method H)
// line 33: C.I undefined (type gno.land/p/filetest/p.C has no method I)

// GnoWordingDiffers:
// line 30: gc expects: ambiguous; gno: C.F undefined (type gno.land/p/filetest/p.C has no method F)
// line 31: gc expects: ambiguous; gno: C.G undefined (type gno.land/p/filetest/p.C has no method G)
// line 32: gc expects: ambiguous; gno: C.H undefined (type gno.land/p/filetest/p.C has no method H)

// Verdict: Clean errorcheck
