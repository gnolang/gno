// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that append arguments requirements are enforced by the
// compiler.

package main

func main() {

	s := make([]int, 8)

	_ = append()           // ERROR "missing arguments to append|not enough arguments for append"
	_ = append(s...)       // ERROR "cannot use ... on first argument|not enough arguments in call to append"
	_ = append(s, 2, s...) // ERROR "too many arguments to append|too many arguments in call to append"

	_ = append(s, make([]int, 0))     // ERROR "cannot use make\(\[\]int, 0\) \(value of type \[\]int\) as int value in argument to append"
	_ = append(s, make([]int, -1)...) // ERROR "negative len argument in make|index -1.* must not be negative"
}

// GoTypeCheckError:
// line 16: invalid operation: not enough arguments for append() (expected 1, found 0)
// line 17: not enough arguments in call to append
// 	have ([]int...)
// 	want ([]int, ...int)
// line 18: too many arguments in call to append
// 	have ([]int, number, []int...)
// 	want ([]int, ...int)
// line 20: cannot use make([]int, 0) (value of type []int) as int value in argument to append
// line 21: invalid argument: index -1 (constant of type int) must not be negative
