// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 20233: panic while formatting an error message

package p

var f = func(...A) // ERROR "undefined: A"

// GnoError:
// line 11: name not defined: A

// GnoWordingDiffers:
// line 11: gc expects: undefined: A; gno: name not defined: A

// Verdict: Clean errorcheck
