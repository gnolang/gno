// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// gc used to recurse infinitely when dowidth is applied
// to a broken recursive type again.
// See golang.org/issue/9432.
package p

type foo struct { // ERROR "invalid recursive type|cycle"
	bar  foo
	blah foo
}

// GnoError:
// line 12: 2: invalid recursive type: foo -> foo
// line 13: expected declaration, found bar
// line 14: expected declaration, found blah
// line 15: expected declaration, found '}'

// GoTypeCheckError:
// line 12: invalid recursive type: foo refers to itself

// GnoOverStrictError:
// line 13: expected declaration, found bar
// line 14: expected declaration, found blah
// line 15: expected declaration, found '}'
