// errorcheck

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 5358: incorrect error message when using f(g()) form on ... args.

package main

func f(x int, y ...int) {}

func g() (int, []int)

func main() {
	f(g()) // ERROR "as int value in|incompatible type"
}

// GnoError:
// line 13: function g does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 16: cannot use g() (value of type []int) as int value in argument to f

// GnoOverStrictError:
// line 13: function g does not have a body but is not natively defined (did you build after pulling from the repository?)
