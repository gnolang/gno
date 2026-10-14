// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

func _(s []int) {
        var i, j, k, l int
        _, _, _, _ = i, j, k, l

        for range s {}
        for i = range s {}
        for i, j = range s {}
        for i, j, k /* ERRORx "range clause permits at most two iteration variables|at most 2 expressions" */ = range s {}
        for i, j, k, l /* ERRORx "range clause permits at most two iteration variables|at most 2 expressions" */ = range s {}
}

func _(s chan int) {
        var i, j, k, l int
        _, _, _, _ = i, j, k, l

        for range s {}
        for i = range s {}
        for i, j /* ERRORx `range over .* permits only one iteration variable` */ = range s {}
        for i, j, k /* ERRORx `range over .* permits only one iteration variable|at most 2 expressions` */ = range s {}
        for i, j, k, l /* ERRORx `range over .* permits only one iteration variable|at most 2 expressions` */ = range s {}
}

// GnoError:
// line 14: expected at most 2 expressions

// GoTypeCheckError:
// line 14

// ParseBlocked:
// line 15: gc expects: range clause permits at most two iteration variables|at most 2 expressions
// line 24: gc expects: range over .* permits only one iteration variable
// line 25: gc expects: range over .* permits only one iteration variable|at most 2 expressions
// line 26: gc expects: range over .* permits only one iteration variable|at most 2 expressions

// Verdict: Clean errorcheck
