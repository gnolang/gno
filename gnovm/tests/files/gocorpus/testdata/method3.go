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

// Verdict: KnownIssue run: #5787 method expression on an interface or mixed-receiver type is rejected at preprocess
