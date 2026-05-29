// errorcheck

// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	var a, b = 1    // ERROR "assignment mismatch: 2 variables but 1 value|wrong number of initializations|cannot initialize"
	_ = 1, 2        // ERROR "assignment mismatch: 1 variable but 2 values|number of variables does not match|cannot assign"
	c, d := 1       // ERROR "assignment mismatch: 2 variables but 1 value|wrong number of initializations|cannot initialize"
	e, f := 1, 2, 3 // ERROR "assignment mismatch: 2 variables but 3 values|wrong number of initializations|cannot initialize"
	_, _, _, _, _, _ = a, b, c, d, e, f
}

// GnoError:
// line 10: missing init expr for b<!VPInvalid(0)>
// line 11: Machine.EvalStaticTypeOf(x) expression not yet preprocessed: 2
// line 12: assignment mismatch: 2 variable(s) but 1 value(s)
// line 13: assignment mismatch: 2 variable(s) but 3 value(s)
