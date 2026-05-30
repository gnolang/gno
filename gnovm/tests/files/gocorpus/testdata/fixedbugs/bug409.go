// run

// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Multiple inlined calls to a function that causes
// redundant address loads.

package main

func F(v [2]float64) [2]float64 {
	return [2]float64{v[0], v[1]}
}

func main() {
	a := F([2]float64{1, 2})
	b := F([2]float64{3, 4})
	println(a[0], a[1], b[0], b[1])
}

// GnoOutput:
// 1 2 3 4

// GoOutput:
// +1.000000e+000 +2.000000e+000 +3.000000e+000 +4.000000e+000

// Divergence: TODO: <category>: explain why this divergence is acceptable