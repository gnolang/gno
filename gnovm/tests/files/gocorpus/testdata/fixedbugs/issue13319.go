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
// line 10: name x not declared
// line 12: unexpected newline in argument list; possibly missing comma or )
// line 16: unexpected newline in argument list; possibly missing comma or )

// Verdict: KnownIssue errorcheck: #6344: line 10 is a type error gc never reaches (it stops at the syntax errors) while the harness checks the partial tree; over-strict: lines 10
