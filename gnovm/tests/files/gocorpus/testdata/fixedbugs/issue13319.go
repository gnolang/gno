// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f(int, int) {
    switch x {
    case 1:
        f(1, g()   // ERROR "expecting \)|possibly missing comma or \)"
    case 2:
        f()
    case 3:
        f(1, g()   // ERROR "expecting \)|possibly missing comma or \)"
    }
}

// GnoError:
// line 10: name x not declared
// line 11: expected '}', found 'case'
// line 12: missing ',' before newline in argument list (and 2 more errors)
// line 13: expected '}', found 'case'
// line 14: wrong argument count in call to f<VPBlock(3,0)>
// line 15: expected '}', found 'case'
// line 16: missing ',' before newline in argument list (and 2 more errors)
// line 18: expected declaration, found '}'

// GoTypeCheckError:
// line 12: missing ',' before newline in argument list (and 2 more errors)
// line 16: missing ',' before newline in argument list (and 2 more errors)

// GnoOverStrictError:
// line 10: name x not declared
// line 11: expected '}', found 'case'
// line 13: expected '}', found 'case'
// line 14: wrong argument count in call to f<VPBlock(3,0)>
// line 15: expected '}', found 'case'
// line 18: expected declaration, found '}'
