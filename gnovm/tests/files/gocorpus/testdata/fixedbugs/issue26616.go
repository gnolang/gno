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
// line 9: decl node already predefined! func three() .res.0 int<VPInvalid(0)>, .res.1 int<VPInvalid(0)>, .res.2 int<VPInvalid(0)> {  }
// line 12: decl node already predefined! func three() .res.0 int<VPInvalid(0)>, .res.1 int<VPInvalid(0)>, .res.2 int<VPInvalid(0)> {  }
// line 13: decl node already predefined! func three() .res.0 int<VPInvalid(0)>, .res.1 int<VPInvalid(0)>, .res.2 int<VPInvalid(0)> {  }
// line 19: function three does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 9
// line 12
// line 13
// line 14
// line 15

// GoTypesOnly:
// line 14: gc expects: assignment mismatch: 1 variable but three returns 3 values|multiple-value function call in single-value context|cannot assign
// line 15: gc expects: assignment mismatch: 1 variable but three returns 3 values|single variable set to multiple-value|multiple-value function call in single-value context|cannot initialize

// GnoOverStrictError:
// line 19: function three does not have a body but is not natively defined (did you build after pulling from the repository?)

// Verdict: KnownIssue errorcheck: 1 over-strict line(s), 0 uncaught marker(s)
