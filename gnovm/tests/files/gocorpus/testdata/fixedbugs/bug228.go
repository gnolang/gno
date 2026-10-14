// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f(x int, y ...int) // ok

func g(x int, y float32) (...)	// ERROR "[.][.][.]"

var x ...int;		// ERROR "[.][.][.]|syntax|type"

type T ...int;		// ERROR "[.][.][.]|syntax|type"

// GnoError:
// line 11: expected type, found ')'

// GoTypeCheckError:
// line 11

// ParseBlocked:
// line 13: gc expects: [.][.][.]|syntax|type
// line 15: gc expects: [.][.][.]|syntax|type

// GnoWordingDiffers:
// line 11: gc expects: [.][.][.]; gno: expected type, found ')'

// Verdict: Clean errorcheck
