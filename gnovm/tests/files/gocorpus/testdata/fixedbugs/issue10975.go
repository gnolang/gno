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

// GnoStaticIncomplete: covered 0 of 1 markers (Gno preprocess: 0, go/types guard: 0); Gno's own preprocess flags none (lenient); the rest are caught by neither — a runnable variant may exercise more

// GnoOverStrictError:
// line 17: struct{} does not implement main.I (missing method int)

// KnownIssue:
// TODO: explain the Gno bug (Gno rejects lines gc + go/types accept)
