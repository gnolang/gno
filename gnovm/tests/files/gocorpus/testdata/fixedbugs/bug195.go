// errorcheck -lang=go1.17

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type I1 interface{ I2 } // ERROR "interface"
type I2 int

type I3 interface{ int } // ERROR "interface"

type S struct { // GC_ERROR "invalid recursive type"
	x interface{ S } // GCCGO_ERROR "interface"
}
type I4 interface { // GC_ERROR "invalid recursive type: I4 refers to itself"
	I4 // GCCGO_ERROR "interface"
}

type I5 interface { // GC_ERROR "invalid recursive type I5\n\tLINE:.* I5 refers to I6\n\tLINE+4:.* I6 refers to I5$"
	I6
}

type I6 interface {
	I5 // GCCGO_ERROR "interface"
}

// GnoError:
// line 14: 2: invalid recursive type: S -> S
// line 15: expected declaration, found x
// line 16: expected declaration, found '}'
// line 17: 2: invalid recursive type: I4 -> I4
// line 18: expected declaration, found I4
// line 19: expected declaration, found '}'
// line 21: 2: invalid recursive type: I5 -> I6 -> I5
// line 22: expected declaration, found I6
// line 23: expected declaration, found '}'
// line 26: expected declaration, found I5
// line 27: expected declaration, found '}'

// GoTypeCheckError:
// line 14: invalid recursive type: S refers to itself
// line 17: invalid recursive type: I4 refers to itself
// line 21: invalid recursive type I5

// GnoOverStrictError:
// line 15: expected declaration, found x
// line 16: expected declaration, found '}'
// line 18: expected declaration, found I4
// line 19: expected declaration, found '}'
// line 22: expected declaration, found I6
// line 23: expected declaration, found '}'
// line 26: expected declaration, found I5
// line 27: expected declaration, found '}'

// UncaughtError:
// line 9: uncaught; gc expects: interface
// line 12: uncaught; gc expects: interface
