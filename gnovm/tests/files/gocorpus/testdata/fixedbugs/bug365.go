// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// check that compiler doesn't stop reading struct def
// after first unknown type.

// Fixes issue 2110.

package main

type S struct {
	err foo.Bar // ERROR "undefined|expected package"
	Num int
}

func main() {
	s := S{}
	_ = s.Num // no error here please
}

// GnoError:
// line 14: 2: name foo not defined in fileset with files [bug365.go]
// line 15: expected declaration, found err
// line 16: expected declaration, found Num
// line 17: expected declaration, found '}'
// line 20: S<VPInvalid(0)> is not a type

// GoTypeCheckError:
// line 15: undefined: foo

// GnoOverStrictError:
// line 14: 2: name foo not defined in fileset with files [bug365.go]
// line 16: expected declaration, found Num
// line 17: expected declaration, found '}'
// line 20: S<VPInvalid(0)> is not a type
