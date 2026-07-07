// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Checking that line number is correct in error message.

package main

type Cint int

func foobar(*Cint, Cint, Cint, *Cint)

func main() {
	a := Cint(1)

	foobar(
		&a,
		0,
		0,
		42, // ERROR ".*"
	)
}

// GnoOverStrictError:
// line 13: function foobar does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 22: cannot use 42 (untyped int constant) as *Cint value in argument to foobar

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)
