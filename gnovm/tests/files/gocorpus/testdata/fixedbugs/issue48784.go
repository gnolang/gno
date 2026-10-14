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
// line 18
// line 19
// line 20
// line 21
// line 22

// GnoWordingDiffers:
// line 18: gc expects: invalid operation:.*mismatched types string and (untyped )?nil; gno: invalid operation: (mismatched types <nil> and string)
// line 19: gc expects: invalid operation:.*mismatched types bool and (untyped )?nil; gno: invalid operation: (mismatched types <nil> and bool)
// line 20: gc expects: invalid operation:.*mismatched types int and (untyped )?nil; gno: invalid operation: (mismatched types <nil> and int)
// line 21: gc expects: invalid operation:.*mismatched types T and (untyped )?nil; gno: invalid operation: (mismatched types <nil> and gno.land/p/filetest/p.T)
// line 22: gc expects: invalid operation:.*mismatched types \[1\]int and (untyped )?nil; gno: invalid operation: (mismatched types <nil> and [1]int)

// Verdict: Clean errorcheck
