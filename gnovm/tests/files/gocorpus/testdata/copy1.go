// errorcheck

// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Verify that copy arguments requirements are enforced by the
// compiler.

package main

func main() {

	si := make([]int, 8)
	sf := make([]float64, 8)

	_ = copy()        // ERROR "not enough arguments"
	_ = copy(1, 2, 3) // ERROR "too many arguments"

	_ = copy(si, "hi") // ERROR "have different element types(.*int.*string| int and byte)"
	_ = copy(si, sf)   // ERROR "have different element types.*int.*float64"

	_ = copy(1, 2)  // ERROR "must be slices; have int, int|expects slice arguments"
	_ = copy(1, si) // ERROR "first argument to copy should be|expects slice arguments"
	_ = copy(si, 2) // ERROR "second argument to copy should be|expects slice arguments"

}

// GnoError:
// line 17: wrong argument count in call to (const (copy func(<X>{}, <Y>{}) int))
// line 18: wrong argument count in call to (const (copy func(<X>{}, <Y>{}) int))

// GoTypeCheckError:
// line 17: invalid operation: not enough arguments for copy() (expected 2, found 0)
// line 18: invalid operation: too many arguments for copy(1, 2, 3) (expected 2, found 3)
// line 20: invalid copy: arguments si (variable of type []int) and "hi" (untyped string constant) have different element types int and byte
// line 21: invalid copy: arguments si (variable of type []int) and sf (variable of type []float64) have different element types int and float64
// line 23: invalid copy: argument must be a slice
// line 24: invalid copy: argument must be a slice
// line 25: invalid copy: argument must be a slice
