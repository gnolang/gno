// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func main() {
	some /* ERROR "undefined" */ [int, int]()
}

type N[T any] struct{}

var _ N [] // ERROR "expected type argument list"

type I interface {
	~[]int
}

func _[T I](i, j int) {
	var m map[int]int
	_ = m[i, j /* ERROR "more than one index" */ ]

	var a [3]int
	_ = a[i, j /* ERROR "more than one index" */ ]

	var s []int
	_ = s[i, j /* ERROR "more than one index" */ ]

	var t T
	_ = t[i, j /* ERROR "more than one index" */ ]
}

// GnoError:
// line 8: invalid operation: more than one index
// line 13: expected type argument list

// ParseBlocked:
// line 21: gc expects: more than one index
// line 24: gc expects: more than one index
// line 27: gc expects: more than one index
// line 30: gc expects: more than one index

// GnoWordingDiffers:
// line 8: gc expects: undefined; gno: invalid operation: more than one index

// Verdict: KnownIssue errorcheck: #6344 uncaught markers: lines 21, 24, 27, 30
