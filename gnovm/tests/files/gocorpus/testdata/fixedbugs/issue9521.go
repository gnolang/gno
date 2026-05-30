// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Test that an incorrect use of the blank identifier is caught.
// Does not compile.

package main

func f() (_, _ []int)         { return }
func g() (x []int, y float64) { return }

func main() {
	_ = append(f()) // ERROR "cannot use \[\]int value as type int in append|cannot use.*type \[\]int.*to append"
	_ = append(g()) // ERROR "cannot use float64 value as type int in append|cannot use.*type float64.*to append"
}

// GnoError:
// line 16: cannot use []int as int
// line 17: cannot use float64 as int

// GoTypeCheckError:
// line 16: cannot use f() (value of type []int) as int value in argument to append
// line 17: cannot use g() (value of type float64) as int value in argument to append
