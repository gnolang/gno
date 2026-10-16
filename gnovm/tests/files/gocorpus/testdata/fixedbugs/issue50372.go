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

// ParseBlocked:
// line 17: gc expects: range clause permits at most two iteration variables
// line 26: gc expects: range over .* permits only one iteration variable
// line 27: gc expects: range over .* permits only one iteration variable
// line 28: gc expects: range over .* permits only one iteration variable

// GnoWordingDiffers:
// line 16: gc expects: range clause permits at most two iteration variables; gno: expected at most 2 expressions

// Verdict: Clean errorcheck
