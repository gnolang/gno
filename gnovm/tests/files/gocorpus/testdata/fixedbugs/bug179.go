// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
L:
	for {
		for {
			break L2    // ERROR "L2"
			continue L2 // ERROR "L2"
		}
	}

L1:
	x := 1
	_ = x
	for {
		break L1    // ERROR "L1"
		continue L1 // ERROR "L1"
	}

	goto L
}

// GnoError:
// line 13: label L2 undefined (and 1 more errors)
// line 14: label L2 undefined
// line 22: cannot find branch label "L1"
// line 23: cannot find branch label "L1"

// GoTypeCheckError:
// line 13: invalid break label L2
// line 14: invalid continue label L2
// line 22: invalid break label L1
// line 23: invalid continue label L1
