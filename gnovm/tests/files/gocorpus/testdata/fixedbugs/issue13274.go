// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Check that we don't ignore EOF.

package p

var f = func() { // ERROR "unexpected EOF|expected .*}.*"

// GnoError:
// line 13: expected '(', found main

// GoTypeCheckError:
// line 13

// UncaughtError:
// line 11: uncaught; gc expects: unexpected EOF|expected .*}.*

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 1 uncaught marker(s)
