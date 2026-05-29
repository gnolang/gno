// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify simple assignment errors are caught by the compiler.
// Does not compile.

package main

import (
	"sync"
	"time"
)

type T struct {
	int
	sync.Mutex
}

func main() {
	{
		var x, y sync.Mutex
		x = y // ok
		_ = x
	}
	{
		var x, y T
		x = y // ok
		_ = x
	}
	{
		var x, y [2]sync.Mutex
		x = y // ok
		_ = x
	}
	{
		var x, y [2]T
		x = y // ok
		_ = x
	}
	{
		x := time.Time{0, 0, nil} // ERROR "assignment.*Time"
		_ = x
	}
	{
		x := sync.Mutex{key: 0} // ERROR "(unknown|assignment).*Mutex"
		_ = x
	}
	{
		x := &sync.Mutex{} // ok
		var y sync.Mutex   // ok
		y = *x             // ok
		*x = y             // ok
		_ = x
		_ = y
	}
	{
		var x = 1
		{
			x, x := 2, 3 // ERROR ".*x.* repeated on left side of :=|x redeclared in this block"
			_ = x
		}
		_ = x
	}
	{
		a, a := 1, 2 // ERROR ".*a.* repeated on left side of :=|a redeclared in this block"
		_ = a
	}
}

// GnoIncomplete: covered 3 of 4 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// GoTypeCheckError:
// line 44: implicit assignment to unexported field sec in struct literal of type time.Time
// line 62: x repeated on left side of :=
// line 68: a repeated on left side of :=

// KnownIssue:
// line 13: unknown import path sync
