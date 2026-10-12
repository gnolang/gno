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

// GoTypeCheckError:
// line 13
// line 14
// line 18
// line 22
// line 23

// GoTypesOnly:
// line 14: gc expects: L2
// line 22: gc expects: L1
// line 23: gc expects: L1

// Verdict: Clean errorcheck
