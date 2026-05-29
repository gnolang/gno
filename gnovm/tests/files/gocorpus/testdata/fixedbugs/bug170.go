// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main
var v1 = ([10]int)(nil);	// ERROR "illegal|nil|invalid"
var v2 [10]int = nil;		// ERROR "illegal|nil|incompatible"
var v3 [10]int;
var v4 = nil;	// ERROR "nil"
func main() {
	v3 = nil;		// ERROR "illegal|nil|incompatible"
}

// GnoError:
// line 8: cannot convert (const (undefined)) to ArrayKind
// line 9: cannot use nil as [10]int value in variable declaration
// line 11: use of untyped nil in variable declaration
// line 13: cannot use nil as [10]int value in assignment

// GoTypeCheckError:
// line 8: cannot convert nil to type [10]int
// line 9: cannot use nil as [10]int value in variable declaration
// line 11: use of untyped nil in variable declaration
// line 13: cannot use nil as [10]int value in assignment
