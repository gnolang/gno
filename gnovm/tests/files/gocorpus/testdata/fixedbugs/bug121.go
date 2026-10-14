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

// GoTypeCheckError:
// line 12

// UncaughtError:
// line 16: uncaught; gc expects: syntax|signature

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 1 uncaught marker(s)
