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
// line 18: should not happen
// line 19: should not happen

// Verdict: KnownIssue errorcheck: #6344; uncaught markers: lines 18, 19; internal errors: lines 18, 19
