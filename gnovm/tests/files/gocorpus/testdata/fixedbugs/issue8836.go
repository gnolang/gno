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

// GnoError:
// line 19: expected block type value path but got VPInvalid:
// --- preprocess stack ---
// stack 2: func main/issue8836.go:15:1-24:2
// ------------------------

// GoTypeCheckError:
// line 22

// GoTypesOnly:
// line 22: gc expects: .*

// GnoOverStrictError:
// line 19: expected block type value path but got VPInvalid:
// --- preprocess stack ---
// stack 2: func main/issue8836.go:15:1-24:2
// ------------------------

// Verdict: KnownIssue errorcheck: 1 over-strict line(s), 0 uncaught marker(s)
