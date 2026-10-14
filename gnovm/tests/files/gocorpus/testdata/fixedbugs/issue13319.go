// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f(int, int) {
    switch x {
    case 1:
        f(1, g()   // ERROR "expecting \)|possibly missing comma or \)"
    case 2:
        f()
    case 3:
        f(1, g()   // ERROR "expecting \)|possibly missing comma or \)"
    }
}

// GnoError:
// line 12: missing ',' before newline in argument list

// GoTypeCheckError:
// line 12

// UncaughtError:
// line 16: uncaught; gc expects: expecting \)|possibly missing comma or \)

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 1 uncaught marker(s)
