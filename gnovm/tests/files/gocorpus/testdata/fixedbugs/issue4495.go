// run

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type I interface {
	m() int
}

type T struct{}

func (T) m() int {
	return 3
}

var t T

var ret = I.m(t)

func main() {
	if ret != 3 {
		println("ret = ", ret)
		panic("ret != 3")
	}
}


// Tracked: issue #5787 (method expressions: interface/promoted/mixed-receiver forms); broken on master, no PR yet.

// GnoOutput:

// GnoError:
// main/issue4495.go:21:11-14: unknown *DeclaredType method named m

// GoOutput:

// KnownIssue:
// Method expressions on interface types are unsupported: I.m(t) is
// rejected at preprocess ("unknown *DeclaredType method named m"). Same
// root cause as fixedbugs/issue29304.go.
