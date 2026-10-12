// errorcheck

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type I interface {
	m(map[I] bool);	// ok
}

type S struct {
	m map[S] bool;	// ERROR "map key type"
}

// GoTypeCheckError:
// line 14

// GoTypesOnly:
// line 14: gc expects: map key type

// Verdict: Clean errorcheck
