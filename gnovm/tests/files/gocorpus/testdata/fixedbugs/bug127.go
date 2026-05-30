// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main
func main() {
	var x int64 = 0;
	println(x != nil);	// ERROR "illegal|incompatible|nil"
	println(0 != nil);	// ERROR "illegal|incompatible|nil"
}

// GnoError:
// line 10: invalid operation: (mismatched types <nil> and int64)
// line 11: invalid operation: (mismatched types <nil> and <untyped> bigint)

// GoTypeCheckError:
// line 10: invalid operation: x != nil (mismatched types int64 and untyped nil)
// line 11: invalid operation: 0 != nil (mismatched types untyped int and untyped nil)
