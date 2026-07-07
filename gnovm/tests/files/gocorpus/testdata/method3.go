// run

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test methods on slices.

package main

type T []int

func (t T) Len() int { return len(t) }

type I interface {
	Len() int
}

func main() {
	var t T = T{0, 1, 2, 3, 4}
	var i I
	i = t
	if i.Len() != 5 {
		println("i.Len", i.Len())
		panic("fail")
	}
	if T.Len(t) != 5 {
		println("T.Len", T.Len(t))
		panic("fail")
	}
	if (*T).Len(&t) != 5 {
		println("(*T).Len", (*T).Len(&t))
		panic("fail")
	}
}


// Tracked: issue #5787 (method expressions: interface/promoted/mixed-receiver forms); broken on master, no PR yet.

// GnoOutput:

// GnoError:
// main/method3.go:31:5-17: cannot use *main.T as []int

// GoOutput:

// KnownIssue:
// Mixed-receiver method expressions are unsupported: (*T).Len(&t) with a
// value-receiver method mistypes the receiver ("cannot use *main.T as
// []int"). Same root cause as method.go and fixedbugs/issue29304.go.
