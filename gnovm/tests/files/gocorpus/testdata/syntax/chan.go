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

// GnoStaticIncomplete: covered 2 of 3 markers (Gno preprocess: 2, go/types guard: 2); Gno bailed before the rest — a runnable variant may exercise more

// GnoError:
// line 11: expected type, found '}' (and 2 more errors)
// line 13: expected '(', found Foo (and 1 more errors)

// GoTypeCheckError:
// line 11: expected type, found '}' (and 2 more errors)
// line 13: expected '(', found Foo (and 1 more errors)
