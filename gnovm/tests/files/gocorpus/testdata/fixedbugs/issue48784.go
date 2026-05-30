// errorcheck -e

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

type T struct{}

var s string
var b bool
var i int
var t T
var a [1]int

var (
	_ = s == nil // ERROR "invalid operation:.*mismatched types string and (untyped )?nil"
	_ = b == nil // ERROR "invalid operation:.*mismatched types bool and (untyped )?nil"
	_ = i == nil // ERROR "invalid operation:.*mismatched types int and (untyped )?nil"
	_ = t == nil // ERROR "invalid operation:.*mismatched types T and (untyped )?nil"
	_ = a == nil // ERROR "invalid operation:.*mismatched types \[1\]int and (untyped )?nil"
)

// GnoError:
// line 18: invalid operation: (mismatched types <nil> and string)
// line 19: invalid operation: (mismatched types <nil> and bool)
// line 20: invalid operation: (mismatched types <nil> and int)
// line 21: invalid operation: (mismatched types <nil> and gno.land/p/filetest/p.T)
// line 22: invalid operation: (mismatched types <nil> and [1]int)

// GoTypeCheckError:
// line 18: invalid operation: s == nil (mismatched types string and untyped nil)
// line 19: invalid operation: b == nil (mismatched types bool and untyped nil)
// line 20: invalid operation: i == nil (mismatched types int and untyped nil)
// line 21: invalid operation: t == nil (mismatched types T and untyped nil)
// line 22: invalid operation: a == nil (mismatched types [1]int and untyped nil)
