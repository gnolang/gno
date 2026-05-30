// errorcheck

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

func printmany(nums ...int) {
	for i, n := range nums {
		fmt.Printf("%d: %d\n", i, n)
	}
	fmt.Printf("\n")
}

func main() {
	printmany(1, 2, 3)
	printmany([]int{1, 2, 3}...)
	printmany(1, "abc", []int{2, 3}...) // ERROR "too many arguments in call( to printmany\n\thave \(number, string, \.\.\.int\)\n\twant \(...int\))?"
}

// GnoError:
// line 21: not enough arguments in call to printmany<VPBlock(3,0)>

// GoTypeCheckError:
// line 21: too many arguments in call to printmany
// 	have (number, string, []int...)
// 	want (...int)
