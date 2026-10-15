// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 4545: untyped constants are incorrectly coerced
// to concrete types when used in interface{} context.

package main

import "fmt"

func main() {
	var s uint
	fmt.Println(1.0 + 1<<s) // ERROR "invalid operation|non-integer type|incompatible type"
	x := 1.0 + 1<<s         // ERROR "invalid operation|non-integer type"
	_ = x
}

// GnoError:
// line 17: operator << not defined on: BigdecKind

// GnoWordingDiffers:
// line 17: gc expects: invalid operation|non-integer type; gno: operator << not defined on: BigdecKind

// Verdict: KnownIssue errorcheck: #6344; uncaught markers: lines 16
