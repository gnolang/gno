// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var _ = int8(4) * 300         // ERROR "overflows int8"
var _ = complex64(1) * 1e200  // ERROR "complex real part overflow|overflows complex64"
var _ = complex128(1) * 1e500 // ERROR "complex real part overflow|overflows complex128"

// GnoError:
// line 9: bigint overflows target kind
// line 10: name complex64 not defined in fileset with files [issue11590.go]
// line 11: name complex128 not defined in fileset with files [issue11590.go]

// GoTypeCheckError:
// line 9: 300 (untyped int constant) overflows int8
// line 10: 1e200 (untyped float constant 1e+200) overflows complex64
// line 11: 1e500 (untyped float constant 1e+500) overflows complex128
