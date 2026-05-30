// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var x int

func main() {
	(x) := 0  // ERROR "non-name [(]x[)]|non-name on left side"
}

// GnoError:
// line 12: no new variables on left side of :=

// GoTypeCheckError:
// line 12: non-name (x) on left side of :=
