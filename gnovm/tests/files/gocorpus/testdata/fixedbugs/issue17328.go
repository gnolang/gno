// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	i := 0
	for ; ; i++) { // ERROR "unexpected \), expecting { after for clause|expected .*{.*|expected .*;.*"
	}
} // GCCGO_ERROR "expected declaration"

// GnoError:
// line 11: expected '{', found ')'
// line 13: expected '}', found 'EOF'

// Verdict: KnownIssue errorcheck: #6344 over-strict: lines 13
