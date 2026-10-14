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
// line 30: unknown *DeclaredType method named F
// line 31: unknown *DeclaredType method named G
// line 32: unknown *DeclaredType method named H
// line 33: unknown *DeclaredType method named I

// GoTypeCheckError:
// line 30
// line 31
// line 32
// line 33

// GnoWordingDiffers:
// line 30: gc expects: ambiguous; gno: unknown *DeclaredType method named F
// line 31: gc expects: ambiguous; gno: unknown *DeclaredType method named G
// line 32: gc expects: ambiguous; gno: unknown *DeclaredType method named H
// line 33: gc expects: no method .*I.*|C.I undefined; gno: unknown *DeclaredType method named I

// Verdict: Clean errorcheck
