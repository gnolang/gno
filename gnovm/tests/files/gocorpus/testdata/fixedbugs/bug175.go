// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f() (int, bool) { return 0, true }

func main() {
	x, y := f(), 2 // ERROR "multi|2-valued"
	_, _ = x, y
}

// GnoError:
// line 12: multiple-value f<VPBlock(3,0)> (value of type [int bool]) in single-value context

// GoTypeCheckError:
// line 12: multiple-value f() (value of type (int, bool)) in single-value context
