// errorcheck -lang=go1.17

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 10975: Returning an invalid interface would cause
// `internal compiler error: getinarg: not a func`.

package main

type I interface {
	int // ERROR "interface contains embedded non-interface|embedding non-interface type"
}

func New() I {
	return struct{}{}
}

// GnoError:
// line 16: 2: [function "New" does not terminate]
// line 17: struct{} does not implement main.I (missing method int)
// line 18: expected declaration, found '}'

// GnoOverStrictError:
// line 16: 2: [function "New" does not terminate]
// line 17: struct{} does not implement main.I (missing method int)
// line 18: expected declaration, found '}'

// UncaughtError:
// line 13: uncaught; gc expects: interface contains embedded non-interface|embedding non-interface type
