// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var c chan int
var v int

func main() {
	if c <- v { // ERROR "cannot use c <- v as value|send statement used as value"
	}
}

var _ = c <- v // ERROR "unexpected <-|send statement used as value"

// GnoStaticIncomplete: covered 1 of 2 markers (Gno preprocess: 1, go/types guard: 1); Gno bailed before the rest — a runnable variant may exercise more

// GnoError:
// line 13: expected boolean expression, found simple statement (missing parentheses around composite literal?) (and 1 more errors)

// GoTypeCheckError:
// line 13: expected boolean expression, found simple statement (missing parentheses around composite literal?) (and 1 more errors)
