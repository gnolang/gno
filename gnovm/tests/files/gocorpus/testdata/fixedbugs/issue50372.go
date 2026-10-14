// errorcheck

// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _(s []int) {
        var i, j, k, l int
        _, _, _, _ = i, j, k, l

        for range s {}
        for i = range s {}
        for i, j = range s {}
        for i, j, k = range s {} // ERROR "range clause permits at most two iteration variables"
        for i, j, k, l = range s {} // ERROR "range clause permits at most two iteration variables"
}

func _(s chan int) {
        var i, j, k, l int
        _, _, _, _ = i, j, k, l

        for range s {}
        for i = range s {}
        for i, j = range s {} // ERROR "range over .* permits only one iteration variable"
        for i, j, k = range s {} // ERROR "range over .* permits only one iteration variable"
        for i, j, k, l = range s {} // ERROR "range over .* permits only one iteration variable"
}

// GnoError:
// line 16: expected at most 2 expressions

// GoTypeCheckError:
// line 16

// UncaughtError:
// line 17: uncaught; gc expects: range clause permits at most two iteration variables
// line 26: uncaught; gc expects: range over .* permits only one iteration variable
// line 27: uncaught; gc expects: range over .* permits only one iteration variable
// line 28: uncaught; gc expects: range over .* permits only one iteration variable

// Verdict: KnownIssue errorcheck: 0 over-strict line(s), 4 uncaught marker(s)
