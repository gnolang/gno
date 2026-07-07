// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that illegal function signatures are detected.
// Does not compile.

package main

type t1 int
type t2 int
type t3 int

func f1(*t2, x t3)	// ERROR "missing parameter name"
func f2(t1, *t2, x t3)	// ERROR "missing parameter name"
func f3() (x int, *string)	// ERROR "missing parameter name"

func f4() (t1 t1)	// legal - scope of parameter named t1 starts in body of f4.

// GnoError:
// line 16: missing parameter name (and 2 more errors)
// line 17: missing parameter name (and 1 more errors)
// line 18: missing parameter name
// line 20: function f4 does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 16: missing parameter name (and 2 more errors)
// line 17: missing parameter name (and 1 more errors)
// line 18: missing parameter name

// GnoOverStrictError:
// line 20: function f4 does not have a body but is not natively defined (did you build after pulling from the repository?)
