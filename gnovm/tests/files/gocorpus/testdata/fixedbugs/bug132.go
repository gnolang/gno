// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T struct {
	x, x int  // ERROR "duplicate|redeclared"
}

// GnoError:
// line 10: x redeclared in this block
// 	previous declaration at bug132.go:10:2

// GoTypeCheckError:
// line 10: x redeclared
