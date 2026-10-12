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
// line 11: illegal character U+0007 (and 4 more errors)

// GoTypeCheckError:
// line 11

// UncaughtError:
// line 12: uncaught; gc expects: illegal character|invalid character
// line 13: uncaught; gc expects: illegal character|invalid character
// line 14: uncaught; gc expects: illegal character|invalid character

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 3 uncaught marker(s)
