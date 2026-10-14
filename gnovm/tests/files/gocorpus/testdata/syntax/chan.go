// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type xyz struct {
    ch chan
} // ERROR "unexpected .*}.* in channel type|missing channel element type"

func Foo(y chan) { // ERROR "unexpected .*\).* in channel type|missing channel element type"
}

func Bar(x chan, y int) { // ERROR "unexpected comma in channel type|missing channel element type"
}

// GnoError:
// line 11: expected type, found '}'

// GoTypeCheckError:
// line 11

// ParseBlocked:
// line 13: gc expects: unexpected .*\).* in channel type|missing channel element type
// line 16: gc expects: unexpected comma in channel type|missing channel element type

// GnoWordingDiffers:
// line 11: gc expects: unexpected .*}.* in channel type|missing channel element type; gno: expected type, found '}'

// Verdict: Clean errorcheck
