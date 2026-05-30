// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	var i int = 100
	i = i << -3 // ERROR "overflows|negative"
}

/*
ixedbugs/bug016.go:7: overflow converting constant to <uint32>UINT32
fixedbugs/bug016.go:7: illegal types for operand: AS
	(<int32>INT32)
*/

// GnoError:
// line 11: invalid operation: negative shift count: (const (-3 <untyped> bigint))

// GoTypeCheckError:
// line 11: invalid operation: negative shift count -3 (untyped int constant)
