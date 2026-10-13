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
// line 13: expected type argument list

// GoTypeCheckError:
// line 13

// UncaughtError:
// line 8: uncaught; gc expects: undefined
// line 21: uncaught; gc expects: more than one index
// line 24: uncaught; gc expects: more than one index
// line 27: uncaught; gc expects: more than one index
// line 30: uncaught; gc expects: more than one index

// KnownIssue:
// errorcheck: 0 over-strict line(s), 5 uncaught marker(s)
