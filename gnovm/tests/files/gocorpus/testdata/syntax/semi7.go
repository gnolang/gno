// errorcheck

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	if x { }	// GCCGO_ERROR "undefined"
	else { }	// ERROR "unexpected semicolon or newline before .?else.?|unexpected keyword else"
}

// GnoError:
// line 10: name x not declared
// line 11: expected statement, found 'else'

// GnoWordingDiffers:
// line 11: gc expects: unexpected semicolon or newline before .?else.?|unexpected keyword else; gno: expected statement, found 'else'

// Verdict: Clean errorcheck
