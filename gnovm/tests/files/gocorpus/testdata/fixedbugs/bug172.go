// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f() {
	a := true;
	a |= a;	// ERROR "illegal.*OR|bool|expected"
}

// GnoError:
// line 11: operator |= not defined on: BoolKind

// GoTypeCheckError:
// line 11

// GnoWordingDiffers:
// line 11: gc expects: illegal.*OR|bool|expected; gno: operator |= not defined on: BoolKind

// Verdict: Clean errorcheck
