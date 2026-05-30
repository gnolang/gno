// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 2520

package main
func main() {
	if 2e9 { }      // ERROR "2e.09|expected bool|non-boolean condition in if statement"
	if 3.14+1i { }  // ERROR "3.14 . 1i|expected bool|non-boolean condition in if statement"
}

// GnoError:
// line 11: cannot convert untyped bigdec to bool
// line 12: imaginaries are not supported

// GoTypeCheckError:
// line 11: non-boolean condition in if statement
// line 12: non-boolean condition in if statement
