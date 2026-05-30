// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	const a uint64 = 10
	var _ int64 = a // ERROR "convert|cannot|incompatible"
}

// GnoError:
// line 11: cannot use uint64 as int64

// GoTypeCheckError:
// line 11: cannot use a (constant 10 of type uint64) as int64 value in variable declaration
