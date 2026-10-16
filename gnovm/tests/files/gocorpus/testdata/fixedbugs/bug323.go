// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type T struct{}
type P *T

func (t *T) Meth() {}
func (t T) Meth2() {}

func main() {
	t := &T{}
	p := P(t)
	p.Meth()  // ERROR "undefined"
	p.Meth2() // ERROR "undefined"
}

// GnoError:
// line 18: missing field Meth in main.P
// line 19: missing field Meth2 in main.P

// Verdict: KnownIssue errorcheck: #6344; uncaught markers: lines 18, 19; internal errors: lines 18, 19
