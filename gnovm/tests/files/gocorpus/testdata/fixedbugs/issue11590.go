// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var _ = int8(4) * 300         // ERROR "overflows int8"
var _ = complex64(1) * 1e200  // ERROR "complex real part overflow|overflows complex64"
var _ = complex128(1) * 1e500 // ERROR "complex real part overflow|overflows complex128"

// GnoTrim: lines 10-11

// GnoTrimmed: lines 10, 11

// GnoError:
// line 9: bigint overflows target kind

// GnoWordingDiffers:
// line 9: gc expects: overflows int8; gno: bigint overflows target kind

// Verdict: Clean errorcheck
