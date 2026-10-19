// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main
func f(a T) T { return a }	// ERROR "undefined"
func main() {
	x := f(0);
	_ = x;
}

// GnoError:
// line 8: name not defined: T

// GnoWordingDiffers:
// line 8: gc expects: undefined; gno: name T not defined in fileset with files [bug198.go]

// Verdict: Clean errorcheck
