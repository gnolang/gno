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

// GnoIncomplete: covered 4 of 5 markers; Gno bailed before the rest — a runnable variant is needed to exercise them

// GnoError:
// line 16: expected at most 2 expressions (and 3 more errors)
// line 17: expected at most 2 expressions (and 2 more errors)
// line 27: expected at most 2 expressions (and 1 more errors)
// line 28: expected at most 2 expressions

// GoTypeCheckError:
// line 16: expected at most 2 expressions (and 3 more errors)
// line 17: expected at most 2 expressions (and 2 more errors)
// line 27: expected at most 2 expressions (and 1 more errors)
// line 28: expected at most 2 expressions
