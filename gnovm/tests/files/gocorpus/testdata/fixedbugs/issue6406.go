// errorcheck

// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	s = "bob" // ERROR "undefined.*s"
	_ = s // ERROR "undefined.*s"
}

// GnoError:
// line 10: name s not declared
// line 11: name s not declared

// GoTypeCheckError:
// line 10
// line 11

// GnoWordingDiffers:
// line 10: gc expects: undefined.*s; gno: name s not declared
// line 11: gc expects: undefined.*s; gno: name s not declared

// Verdict: Clean errorcheck
