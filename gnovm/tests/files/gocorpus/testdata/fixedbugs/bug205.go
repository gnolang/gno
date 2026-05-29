// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var t []int
var s string;
var m map[string]int;

func main() {
	println(t["hi"]); // ERROR "non-integer slice index|must be integer|cannot convert"
	println(s["hi"]); // ERROR "non-integer string index|must be integer|cannot convert"
	println(m[0]);    // ERROR "cannot use.*as type string|cannot convert|cannot use"
}

// GnoError:
// line 14: type should be numeric
// line 15: type should be numeric
// line 16: cannot use untyped Bigint as StringKind

// GoTypeCheckError:
// line 14: cannot convert "hi" (untyped string constant) to type int
// line 15: cannot convert "hi" (untyped string constant) to type int
// line 16: cannot use 0 (untyped int constant) as string value in map index
