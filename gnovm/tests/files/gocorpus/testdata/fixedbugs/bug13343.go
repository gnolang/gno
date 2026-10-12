// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

var (
	a, b = f() // ERROR "initialization cycle|depends upon itself|depend upon each other"
	c    = b   // GCCGO_ERROR "depends upon itself|depend upon each other"
)

func f() (int, int) {
	return c, c
}

func main() {}

// GoTypeCheckError:
// line 10
// line 11
// line 14

// GoTypesOnly:
// line 10: gc expects: initialization cycle|depends upon itself|depend upon each other

// Verdict: Clean errorcheck
