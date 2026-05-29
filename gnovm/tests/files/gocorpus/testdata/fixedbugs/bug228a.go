// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f(x int, y ...int) // ok

func h(x, y ...int) // ERROR "[.][.][.]"

func i(x int, y ...int, z float32) // ERROR "[.][.][.]"

// GnoIncomplete: covered 0 of 2 markers; Gno bailed before the rest — a runnable variant is needed to exercise them
// GnoError:
// line 9: function f does not have a body but is not natively defined (did you build after pulling from the repository?)
