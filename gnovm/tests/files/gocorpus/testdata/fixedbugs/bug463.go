// errorcheck

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Issue 3757: unhelpful typechecking loop message
// for constants that refer to themselves.

package main

const a = a // ERROR "refers to itself|definition loop|initialization cycle"

const (
	X    = A
	A    = B // ERROR "refers to itself|definition loop|initialization cycle"
	B    = D
	C, D = 1, A
)

func main() {
}

// GnoError:
// line 10: decl node already predefined! const A<!VPInvalid(0)> = B<VPInvalid(0)>:
// --- preprocess stack ---
// stack 0: file main/bug463.go:10:1-22:2
// ------------------------

// Verdict: KnownIssue errorcheck: #6344; uncaught markers: lines 12, 16; internal errors: lines 10; on-chain path first error on line 12, not reported by accumulate mode
