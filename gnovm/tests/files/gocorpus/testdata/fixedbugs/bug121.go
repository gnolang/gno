// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T func()

type I interface {
	f, g ();  // ERROR "unexpected comma"
}

type J interface {
	h T;  // ERROR "syntax|signature"
}

// GnoError:
// line 12: expected ';', found ','
// line 16: expected ';', found T
// line 17: expected ';', found 'EOF'

// Verdict: KnownIssue errorcheck: #6344 over-strict: lines 17
