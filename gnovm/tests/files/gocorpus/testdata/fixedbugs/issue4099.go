// errorcheck -0 -m

//go:build !goexperiment.newinliner

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check go:noescape annotations.

package p

// The noescape comment only applies to the next func,
// which must not have a body.

//go:noescape

func F1([]byte)

func F2([]byte)

func G() {
	var buf1 [10]byte
	F1(buf1[:])

	var buf2 [10]byte // ERROR "moved to heap: buf2"
	F2(buf2[:])
}

// GnoIncomplete: covered 0 of 1 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// KnownIssue:
// line 18: function F1 does not have a body but is not natively defined (did you build after pulling from the repository?)
