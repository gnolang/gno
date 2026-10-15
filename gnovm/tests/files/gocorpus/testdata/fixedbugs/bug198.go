// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main
func f(a T) T { return a }	// ERROR "undefined"
func main() {
	x := f(0);
	_ = x;
}

// GnoError:
// line 8: name T not defined in fileset with files [bug198.go]
// line 11: name x not declared

// Verdict: KnownIssue errorcheck: #6344; over-strict: lines 11
