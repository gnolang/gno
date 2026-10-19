// errorcheck -d=panic

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var m = map[string]int{
	"a": 1,
	1:   1, // ERROR "cannot use 1.*as.*string.*in map"
	2:   2, // ERROR "cannot use 2.*as.*string.*in map"
}

// GnoError:
// line 9: cannot use untyped Bigint as StringKind

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 11, 12; over-strict: lines 9
