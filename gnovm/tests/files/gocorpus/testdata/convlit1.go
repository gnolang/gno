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

// GoTypeCheckError:
// line 12: cannot use "a" (untyped string constant) as int value in array or slice literal
// line 13: invalid composite literal type int
// line 19: cannot convert 1 (untyped int constant) to type func() int

// KnownIssue:
// line 16: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
