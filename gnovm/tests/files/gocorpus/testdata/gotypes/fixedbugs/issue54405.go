// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that we don't see spurious errors for ==
// for values with invalid types due to prior errors.

package p

var x struct {
	f *NotAType /* ERROR "undefined" */
}
var _ = x.f == nil // no error expected here

var y *NotAType  /* ERROR "undefined" */
var _ = y == nil // no error expected here

// GnoError:
// line 10: name NotAType not defined in fileset with files [issue54405.go]
// line 15: name NotAType not defined in fileset with files [issue54405.go]

// GnoWordingDiffers:
// line 15: gc expects: undefined; gno: name NotAType not defined in fileset with files [issue54405.go]

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 11; over-strict: lines 10
