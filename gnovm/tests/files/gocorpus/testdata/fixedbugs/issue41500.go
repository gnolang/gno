// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package p

type s struct {
	slice []int
}

func f() {
	var x *s

	_ = x == nil || len(x.slice) // ERROR "invalid operation: .+ \(operator \|\| not defined on int\)|incompatible types|mismatched types untyped bool and int"
	_ = len(x.slice) || x == nil // ERROR "invalid operation: .+ \(operator \|\| not defined on int\)|incompatible types|mismatched types int and untyped bool"
	_ = x == nil && len(x.slice) // ERROR "invalid operation: .+ \(operator && not defined on int\)|incompatible types|mismatched types untyped bool and int"
	_ = len(x.slice) && x == nil // ERROR "invalid operation: .+ \(operator && not defined on int\)|incompatible types|mismatched types int and untyped bool"
}

// GnoError:
// line 16: operator || not defined on: IntKind
// line 17: operator || not defined on: IntKind
// line 18: operator && not defined on: IntKind
// line 19: operator && not defined on: IntKind

// GnoWordingDiffers:
// line 16: gc expects: invalid operation: .+ \(operator \|\| not defined on int\)|incompatible types|mismatched types untyped bool and int; gno: operator || not defined on: IntKind
// line 17: gc expects: invalid operation: .+ \(operator \|\| not defined on int\)|incompatible types|mismatched types int and untyped bool; gno: operator || not defined on: IntKind
// line 18: gc expects: invalid operation: .+ \(operator && not defined on int\)|incompatible types|mismatched types untyped bool and int; gno: operator && not defined on: IntKind
// line 19: gc expects: invalid operation: .+ \(operator && not defined on int\)|incompatible types|mismatched types int and untyped bool; gno: operator && not defined on: IntKind

// Verdict: Clean errorcheck
