// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that illegal uses of composite literals are detected.
// Does not compile.

package main

var a = []int { "a" };	// ERROR "conver|incompatible|cannot"
var b = int { 1 };	// ERROR "compos"


func f() int

func main() {
	if f < 1 { }	// ERROR "conver|incompatible|invalid"
}

// GnoError:
// line 12: cannot use untyped string as IntKind
// line 13: unexpected composite lit type int
// line 16: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 19: operator < not defined on: FuncKind

// GnoWordingDiffers:
// line 19: gc expects: conver|incompatible|invalid; gno: operator < not defined on: FuncKind

// Verdict: Clean errorcheck
