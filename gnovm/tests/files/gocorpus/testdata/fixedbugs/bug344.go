// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

func main() {
	// invalid use of goto.
	// do whatever you like, just don't crash.
	i := 42
	a := []*int{&i, &i, &i, &i}
	x := a[0]
	goto start  // ERROR "jumps into block"
	z := 1
	_ = z
	for _, x = range a {	// GCCGO_ERROR "block"
	start:
		fmt.Sprint(*x)
	}
}

// GnoError:
// line 17: cannot find GOTO label "start" within current function

// GoTypeCheckError:
// line 17: goto start jumps into block
