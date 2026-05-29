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

// GnoIncomplete: covered 1 of 2 markers; Gno bailed before the rest — a runnable variant is needed to exercise them
// GnoError:
// line 17: invalid operation: shifted operand 1 (type float64) must be integer
