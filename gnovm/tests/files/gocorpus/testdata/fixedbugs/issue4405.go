// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

const (
	_ = iota
	_ // ERROR "illegal character|invalid character"
	_  // ERROR "illegal character|invalid character"
	_  // ERROR "illegal character|invalid character"
	_  // ERROR "illegal character|invalid character"
)

// GnoError:
// line 11: expected ';', found 'ILLEGAL'
// line 12: illegal character U+0008
// line 13: illegal character U+000B
// line 14: illegal character U+000C
// line 15: expected ')', found 'EOF'

// GnoWordingDiffers:
// line 11: gc expects: illegal character|invalid character; gno: expected ';', found 'ILLEGAL'

// Verdict: KnownIssue errorcheck: #6344 over-strict: lines 15
