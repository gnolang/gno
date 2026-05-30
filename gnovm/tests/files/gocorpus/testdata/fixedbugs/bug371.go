// errorcheck

// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 2343

package main

type T struct{}

func (t *T) pm() {}
func (t T) m()   {}

func main() {
	p := &T{}
	p.pm()
	p.m()

	q := &p
	q.m()  // ERROR "requires explicit dereference|undefined"
	q.pm() // ERROR "requires explicit dereference|undefined"
}

// GnoError:
// line 22: missing field m in **main.T
// line 23: missing field pm in **main.T

// GoTypeCheckError:
// line 22: q.m undefined (type **T has no field or method m)
// line 23: q.pm undefined (type **T has no field or method pm)
