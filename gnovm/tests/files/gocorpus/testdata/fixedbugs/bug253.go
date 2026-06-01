// run

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

type S1 struct {
	i int
}
type S2 struct {
	i int
}
type S3 struct {
	S1
	S2
}
type S4 struct {
	S3
	S1
}

func main() {
	var s4 S4
	if s4.i != 0 { // .i refers to s4.S1.i, unambiguously
		panic("fail")
	}
}

// GnoOutput:

// GnoError:
// main/bug253.go:26:5-9: missing field i in main.S4

// GoOutput:

// KnownIssue:
// TODO: explain the Gno bug (Gno errors where Go runs clean)
