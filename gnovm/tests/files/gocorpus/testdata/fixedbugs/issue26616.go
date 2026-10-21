// errorcheck

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package p

var x int = three() // ERROR "assignment mismatch: 1 variable but three returns 3 values|multiple-value function call in single-value context|multiple-value "

func f() {
	var _ int = three() // ERROR "assignment mismatch: 1 variable but three returns 3 values|multiple-value function call in single-value context|multiple-value "
	var a int = three() // ERROR "assignment mismatch: 1 variable but three returns 3 values|multiple-value function call in single-value context|multiple-value "
	a = three()         // ERROR "assignment mismatch: 1 variable but three returns 3 values|multiple-value function call in single-value context|cannot assign"
	b := three()        // ERROR "assignment mismatch: 1 variable but three returns 3 values|single variable set to multiple-value|multiple-value function call in single-value context|cannot initialize"
	_, _ = a, b
}

func three() (int, int, int)

// GnoError:
// line 9: multiple-value three<VPBlock(2,2)> (value of type [int int int]) in single-value context
// line 12: multiple-value three<VPBlock(3,2)> (value of type [int int int]) in single-value context
// line 13: multiple-value three<VPBlock(3,2)> (value of type [int int int]) in single-value context
// line 15: multiple-value three<VPBlock(3,2)> (value of type [int int int]) in single-value context
// line 19: function three does not have a body but is not natively defined (did you build after pulling from the repository?)

// GnoWordingDiffers:
// line 15: gc expects: assignment mismatch: 1 variable but three returns 3 values|single variable set to multiple-value|multiple-value function call in single-value context|cannot initialize; gno: multiple-value three<VPBlock(3,2)> (value of type [int int int]) in single-value context

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 14
