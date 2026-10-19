// errorcheck -d=panic

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 20245: panic while formatting an error message

package p

var e = interface{ I1 } // ERROR "undefined: I1"

// GnoError:
// line 11: name not defined: I1

// GnoWordingDiffers:
// line 11: gc expects: undefined: I1; gno: name not defined: I1

// Verdict: Clean errorcheck
