// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that the correct (not implicitly dereferenced)
// type is reported in the error message.

package p

var p = &[1]complex128{0}
var _ = real(p)  // ERROR "type \*\[1\]complex128|argument must have complex type"
var _ = imag(p)	 // ERROR "type \*\[1\]complex128|argument must have complex type"

// GoTypeCheckError:
// line 13: invalid argument: argument has type *[1]complex128, expected complex type
// line 14: invalid argument: argument has type *[1]complex128, expected complex type

// KnownIssue:
// line 12: name complex128 not defined in fileset with files [issue20602.go]
