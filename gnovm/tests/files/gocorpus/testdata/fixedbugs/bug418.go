// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 3044.
// Multiple valued expressions in return lists.

package p

func Two() (a, b int)

// F used to compile.
func F() (x interface{}, y int) {
	return Two(), 0 // ERROR "single-value context|2\-valued"
}

// Recursive used to trigger an internal compiler error.
func Recursive() (x interface{}, y int) {
	return Recursive(), 0 // ERROR "single-value context|2\-valued"
}

// GoTypeCheckError:
// line 16: multiple-value Two() (value of type (a int, b int)) in single-value context
// line 21: multiple-value Recursive() (value of type (x interface{}, y int)) in single-value context

// KnownIssue:
// line 12: function Two does not have a body but is not natively defined (did you build after pulling from the repository?)
