// errorcheck

// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

func f(a int, b ...int) {}

func main() {
	f(nil...) // ERROR "not enough arguments in call to f\n\thave \(nil\)\n\twant \(int, \[\]int\)|not enough arguments"
}

// GnoError:
// line 12: not enough arguments in call to f<VPBlock(3,0)>

// GoTypeCheckError:
// line 12: not enough arguments in call to f
// 	have (nil...)
// 	want (int, ...int)
