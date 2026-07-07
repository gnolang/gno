// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f(x int, y ...int) // ok

func h(x, y ...int) // ERROR "[.][.][.]"

func i(x int, y ...int, z float32) // ERROR "[.][.][.]"

// GnoError:
// line 9: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 11: function h does not have a body but is not natively defined (did you build after pulling from the repository?)
// line 13: function i does not have a body but is not natively defined (did you build after pulling from the repository?)

// GoTypeCheckError:
// line 11: can only use ... with final parameter (and 1 more errors)
// line 13: can only use ... with final parameter (and 1 more errors)

// GnoOverStrictError:
// line 9: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
