// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that it is illegal to take the address of a function.
// Does not compile.

package main

var notmain func()

func main() {
	var x = &main		// ERROR "address of|invalid"
	main = notmain	// ERROR "assign to|invalid"
	_ = x
}

// GnoError:
// line 16: not assignable

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 15
