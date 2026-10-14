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
// line 21: missing field Num in main.S

// GoTypeCheckError:
// line 15

// GnoOverStrictError:
// line 21: missing field Num in main.S

// GnoWordingDiffers:
// line 15: gc expects: undefined|expected package; gno: 2: name foo not defined in fileset with files [bug365.go]

// Verdict: KnownIssue errorcheck: 1 over-strict line(s), 0 uncaught marker(s)
