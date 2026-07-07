// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 1556
package foo

type I interface {
	m() int
}

type T int

var _ I = T(0)	// GCCGO_ERROR "incompatible"

func (T) m(buf []byte) (a int, b xxxx) {  // ERROR "xxxx"
	return 0, nil
}

// GnoError:
// line 16: gno.land/p/filetest/foo.T does not implement gno.land/p/filetest/foo.I (missing method m)
// line 18: 2: name xxxx not defined in fileset with files [bug374.go]
// line 19: expected declaration, found 'return'
// line 20: expected declaration, found '}'

// GoTypeCheckError:
// line 18: undefined: xxxx

// GnoOverStrictError:
// line 16: gno.land/p/filetest/foo.T does not implement gno.land/p/filetest/foo.I (missing method m)
// line 19: expected declaration, found 'return'
// line 20: expected declaration, found '}'
