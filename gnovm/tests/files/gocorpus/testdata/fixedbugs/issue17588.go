// errorcheck -d=panic

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 17588: internal compiler error in typecheckclosure()
// because in case of Func.Nname.Type == nil, Decldepth
// is not initialized in typecheckfunc(). This test
// produces that case.

package p

type F func(b T) // ERROR "T .*is not a type|expected type"

func T(fn F) {
	func() {
		fn(nil) // If Decldepth is not initialized properly, typecheckclosure() Fatals here.
	}()
}

// GnoError:
// line 14: T<VPBlock(2,1)> is not a type
// line 17: expected 'IDENT', found '{' (and 4 more errors)
// line 18: expected declaration, found fn
// line 19: expected declaration, found '}'
// line 20: expected declaration, found '}'

// GoTypeCheckError:
// line 14: T is not a type

// GnoOverStrictError:
// line 17: expected 'IDENT', found '{' (and 4 more errors)
// line 18: expected declaration, found fn
// line 19: expected declaration, found '}'
// line 20: expected declaration, found '}'
