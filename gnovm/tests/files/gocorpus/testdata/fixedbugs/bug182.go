// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	x := 0;
	if x {	// ERROR "x.*int|bool"
	}
}

// GnoError:
// line 11: 3: expected typed bool kind, but got IntKind

// GoTypeCheckError:
// line 11: non-boolean condition in if statement
