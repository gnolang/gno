// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f() /* no return type */ {}

func main() {
	x := f();  // ERROR "mismatch|as value|no type"
	_ = x
}

// GnoError:
// line 12: f<VPBlock(3,0)> (no value) used as value

// GoTypeCheckError:
// line 12: f() (no value) used as value
