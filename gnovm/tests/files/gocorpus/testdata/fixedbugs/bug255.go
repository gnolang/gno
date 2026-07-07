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
// line 15: name ee not defined in fileset with files [bug255.go]
// line 16: bigint overflows target kind
// line 19: function ff does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 11: array length 1.5 (untyped float constant) must be integer
// line 12: array length "abc" (untyped string constant) must be integer
// line 13: invalid array length nil
// line 15: undefined array length ee or missing type constraint
// line 16: invalid array length 1 << 65 (untyped int constant 36893488147419103232)
// line 21: array length len([1]string{…}) (value of type int) must be constant

// GnoOverStrictError:
// line 19: function ff does not have a body but is not natively defined (did you build after pulling from the repository?)
