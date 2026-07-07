// run

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 4353. An optimizer bug in 8g triggers a runtime fault
// instead of an out of bounds panic.

package main

var aib [100000]int
var paib *[100000]int = &aib
var i64 int64 = 100023

func main() {
	defer func() { recover() }()
	_ = paib[i64]
}


// Fixed: master PR #5738 (1da3a0ff7); verified clean, broken at parent; re-golden after rebase.

// GnoOutput:

// GnoError:
// runtime error: index out of range [100023] with length 100000

// GoOutput:

// KnownIssue:
// Out-of-range index on a pointer-to-array (paib[i64]) escaped as an
// unrecoverable VM panic instead of a recoverable runtime panic, so
// recover() never fired. Same root cause as fixedbugs/issue15252.go.
