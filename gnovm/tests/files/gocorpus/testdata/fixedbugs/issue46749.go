// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var s string
var b bool
var i int
var iface interface{}

var (
	_ = "" + b   // ERROR "invalid operation.*mismatched types.*untyped string and bool"
	_ = "" + i   // ERROR "invalid operation.*mismatched types.*untyped string and int"
	_ = "" + nil // ERROR "invalid operation.*mismatched types.*untyped string and nil|(untyped nil)"
)

var (
	_ = s + false // ERROR "invalid operation.*mismatched types.*string and untyped bool"
	_ = s + 1     // ERROR "invalid operation.*mismatched types.*string and untyped int"
	_ = s + nil   // ERROR "invalid operation.*mismatched types.*string and nil|(untyped nil)"
)

var (
	_ = "" + false // ERROR "invalid operation.*mismatched types.*untyped string and untyped bool"
	_ = "" + 1     // ERROR "invalid operation.*mismatched types.*untyped string and untyped int"
)

var (
	_ = b + 1         // ERROR "invalid operation.*mismatched types.*bool and untyped int"
	_ = i + false     // ERROR "invalid operation.*mismatched types.*int and untyped bool"
	_ = iface + 1     // ERROR "invalid operation.*mismatched types.*interface *{} and int"
	_ = iface + 1.0   // ERROR "invalid operation.*mismatched types.*interface *{} and float64"
	_ = iface + false // ERROR "invalid operation.*mismatched types.*interface *{} and bool"
)

// GnoError:
// line 15: invalid operation: "" + b (mismatched types untyped string and bool)
// line 16: invalid operation: "" + i (mismatched types untyped string and int)
// line 17: invalid operation: (const ("" <untyped> string)) + (const (undefined)) (mismatched types <untyped> string and untyped nil)
// line 21: invalid operation: s + false (mismatched types string and untyped bool)
// line 22: invalid operation: s + 1 (mismatched types string and untyped int)
// line 23: invalid operation: s<VPBlock(2,0)> + (const (undefined)) (mismatched types string and untyped nil)
// line 27: invalid operation: "" + false (mismatched types untyped string and untyped bool)
// line 28: invalid operation: "" + 1 (mismatched types untyped string and untyped int)
// line 32: invalid operation: b + 1 (mismatched types bool and untyped int)
// line 33: invalid operation: i + false (mismatched types int and untyped bool)
// line 34: invalid operation: iface + 1 (mismatched types interface{} and int)
// line 35: invalid operation: iface + 1.0 (mismatched types interface{} and float64)
// line 36: invalid operation: iface + false (mismatched types interface{} and bool)
