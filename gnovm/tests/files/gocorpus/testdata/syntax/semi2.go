// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	switch x; y	// ERROR "missing .*{.* after switch clause|undefined"
	{
		z

// GnoError:
// line 10: expected '{', found newline
// line 11: expected '}', found '{'
// line 12: expected ';', found z

// GnoWordingDiffers:
// line 10: gc expects: missing .*{.* after switch clause|undefined; gno: expected '{', found newline

// Verdict: KnownIssue errorcheck: #6344 over-strict: lines 11, 12
