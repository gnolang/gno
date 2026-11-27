// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sort

// lessSwap stands in for an Interface in the sort routines of zsortfunc.gno.
// They call Less directly and swap the elements of x through sliceSwap.
type lessSwap struct {
	Less func(i, j int) bool
	x    any
}

// Slice sorts the slice x given the provided less function.
// It panics if x is not a slice.
//
// The sort is not guaranteed to be stable: equal elements
// may be reversed from their original order.
// For a stable sort, use SliceStable.
//
// The less function must satisfy the same requirements as
// the Interface type's Less method.
func Slice(x any, less func(i, j int) bool) {
	n := sliceLen(x)
	quickSort_func(&lessSwap{less, x}, 0, n, maxDepth(n))
}

// SliceStable sorts the slice x using the provided less
// function, keeping equal elements in their original order.
// It panics if x is not a slice.
//
// The less function must satisfy the same requirements as
// the Interface type's Less method.
func SliceStable(x any, less func(i, j int) bool) {
	stable_func(&lessSwap{less, x}, sliceLen(x))
}

// SliceIsSorted reports whether the slice x is sorted according to the provided less function.
// It panics if x is not a slice.
func SliceIsSorted(x any, less func(i, j int) bool) bool {
	for i := sliceLen(x) - 1; i > 0; i-- {
		if less(i, i-1) {
			return false
		}
	}
	return true
}

// sliceLen returns the length of the slice x. It panics if x is not a slice.
func sliceLen(x any) int

// sliceSwap swaps x[i] and x[j]. It panics if x is not a slice, if either
// index is out of range, or if the caller may not write to x.
func sliceSwap(x any, i, j int)
