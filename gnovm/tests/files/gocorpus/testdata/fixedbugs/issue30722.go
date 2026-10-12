// errorcheck

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that we only get one error per invalid integer literal.

package p

const (
	_ = 1_       // ERROR "'_' must separate successive digits"
	_ = 0b       // ERROR "binary literal has no digits|invalid numeric literal"
	_ = 0o       // ERROR "octal literal has no digits|invalid numeric literal"
	_ = 0x       // ERROR "hexadecimal literal has no digits|invalid numeric literal"
	_ = 0xde__ad // ERROR "'_' must separate successive digits"
)

// GnoError:
// line 12: '_' must separate successive digits (and 4 more errors)

// GoTypeCheckError:
// line 12

// UncaughtError:
// line 13: uncaught; gc expects: binary literal has no digits|invalid numeric literal
// line 14: uncaught; gc expects: octal literal has no digits|invalid numeric literal
// line 15: uncaught; gc expects: hexadecimal literal has no digits|invalid numeric literal
// line 16: uncaught; gc expects: '_' must separate successive digits

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 4 uncaught marker(s)
