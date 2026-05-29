// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	x := ""
	x = +"hello"  // ERROR "invalid operation.*string|expected numeric"
	x = +x  // ERROR "invalid operation.*string|expected numeric"
}

// GnoError:
// line 11: invalid operation: operator + not defined on "hello" (untyped string constant)
// line 12: invalid operation: operator + not defined on x (variable of type string)
