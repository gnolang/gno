// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that large integer constant expressions cause overflow.
// Does not compile.

package main

const (
	A int = 1
	B byte;	// ERROR "type without expr|expected .=.|missing init expr"
)

const LargeA = 1000000000000000000
const LargeB = LargeA * LargeA * LargeA
const LargeC = LargeB * LargeB * LargeB // GC_ERROR "constant multiplication overflow"

const AlsoLargeA = LargeA << 400 << 400 >> 400 >> 400 // GC_ERROR "constant shift overflow"

// Issue #42732.

const a = 1e+500000000
const b = a * a // ERROR "constant multiplication overflow|not representable"
const c = b * b

const MaxInt512 = (1<<256 - 1) * (1<<256 + 1)
const _ = MaxInt512 + 1  // ERROR "constant addition overflow"
const _ = MaxInt512 ^ -1 // ERROR "constant bitwise XOR overflow"
const _ = ^MaxInt512     // ERROR "constant bitwise complement overflow"

// GnoError:
// line 10: 21: constant shift overflow
// line 14: missing init expr for const declaration
// line 19: constant multiplication overflow
// line 30: constant addition overflow
// line 31: constant bitwise XOR overflow
// line 32: constant bitwise complement overflow

// GoTypeCheckError:
// line 14
// line 19
// line 21
// line 26
// line 30
// line 31
// line 32

// Verdict: Clean errorcheck
