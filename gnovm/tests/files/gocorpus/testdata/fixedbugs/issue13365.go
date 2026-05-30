// errorcheck

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// issue 13365: confusing error message (array vs slice)

package main

var t struct{}

func main() {
	_ = []int{-1: 0}    // ERROR "index must be non\-negative integer constant|index expression is negative|must not be negative"
	_ = [10]int{-1: 0}  // ERROR "index must be non\-negative integer constant|index expression is negative|must not be negative"
	_ = [...]int{-1: 0} // ERROR "index must be non\-negative integer constant|index expression is negative|must not be negative"

	_ = []int{100: 0}
	_ = [10]int{100: 0} // ERROR "index 100 out of bounds|out of range"
	_ = [...]int{100: 0}

	_ = []int{t}    // ERROR "cannot use .* as (type )?int( in slice literal)?|incompatible type"
	_ = [10]int{t}  // ERROR "cannot use .* as (type )?int( in array literal)?|incompatible type"
	_ = [...]int{t} // ERROR "cannot use .* as (type )?int( in array literal)?|incompatible type"
}

// GnoError:
// line 14: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 15: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 16: invalid argument: index must not be negative: (-1 <untyped> bigint)
// line 22: cannot use struct{} as int
// line 23: cannot use struct{} as int
// line 24: cannot use struct{} as int

// GoTypeCheckError:
// line 14: invalid argument: index -1 (constant of type int) must not be negative
// line 15: invalid argument: index -1 (constant of type int) must not be negative
// line 16: invalid argument: index -1 (constant of type int) must not be negative
// line 19: invalid argument: index 100 out of bounds [0:10]
// line 22: cannot use t (variable of type struct{}) as int value in array or slice literal
// line 23: cannot use t (variable of type struct{}) as int value in array or slice literal
// line 24: cannot use t (variable of type struct{}) as int value in array or slice literal
