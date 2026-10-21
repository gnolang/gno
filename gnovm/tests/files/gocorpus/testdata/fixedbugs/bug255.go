// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var a [10]int    // ok
var b [1e1]int   // ok
var c [1.5]int   // ERROR "truncated|must be integer"
var d ["abc"]int // ERROR "invalid array bound|not numeric|must be integer"
var e [nil]int   // ERROR "use of untyped nil|invalid array (bound|length)|not numeric|must be constant"
// var f [e]int  // ok with Go 1.17 because an error was reported for e; leads to an error for Go 1.18
var f [ee]int      // ERROR "undefined|undeclared"
var g [1 << 65]int // ERROR "array bound is too large|overflows|invalid array length"
var h [len(a)]int  // ok

func ff() string

var i [len([1]string{ff()})]int // ERROR "non-constant array bound|not constant|must be constant"

// GnoError:
// line 11: cannot convert untyped bigdec to integer -- 1.5 not an exact integer
// line 12: cannot convert StringKind to IntKind
// line 13: cannot convert (undefined) to int
// line 15: name not defined: ee
// line 16: bigint overflows target kind
// line 19: function ff does not have a body but is not natively defined (did you build after pulling from the repository?)

// GnoWordingDiffers:
// line 11: gc expects: truncated|must be integer; gno: cannot convert untyped bigdec to integer -- 1.5 not an exact integer
// line 12: gc expects: invalid array bound|not numeric|must be integer; gno: cannot convert StringKind to IntKind
// line 13: gc expects: use of untyped nil|invalid array (bound|length)|not numeric|must be constant; gno: cannot convert (undefined) to int
// line 15: gc expects: undefined|undeclared; gno: name not defined: ee

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 21
