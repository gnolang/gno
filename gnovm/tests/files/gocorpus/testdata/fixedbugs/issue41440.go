// errorcheck

// Copyright 2020 The Go Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package p

func f(...int) {}

func g() {
	var x []int
	f(x, x...) // ERROR "have \(\[\]int, \.\.\.int\)|too many arguments"
}

// GnoError:
// line 13: not enough arguments in call to f<VPBlock(3,0)>

// GoTypeCheckError:
// line 13: too many arguments in call to f
// 	have ([]int, []int...)
// 	want (...int)
