// errorcheck -d=panic

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// Issue 2623
var m = map[string]int{
	"abc": 1,
	1:     2, // ERROR "cannot use 1.*as type string in map key|incompatible type|cannot convert|cannot use"
}

// GnoError:
// line 12: cannot use untyped Bigint as StringKind

// Verdict: KnownIssue errorcheck: #6344 on-chain path first error on line 10, not reported by accumulate mode
