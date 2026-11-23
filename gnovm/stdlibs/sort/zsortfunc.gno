// Code generated from sort.gno by zsortfunc_test.go; DO NOT EDIT.
// Regenerate with: go test ./gnovm/stdlibs/sort -run TestZSortFunc -update

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sort

// Auto-generated variant of sort.gno:insertionSort
func insertionSort_func(data lessSwap, a, b int) {
	for i := a + 1; i < b; i++ {
		for j := i; j > a && data.Less(j, j-1); j-- {
			sliceSwap(data.x, j, j-1)
		}
	}
}

// Auto-generated variant of sort.gno:siftDown
func siftDown_func(data lessSwap, lo, hi, first int) {
	root := lo
	for {
		child := 2*root + 1
		if child >= hi {
			break
		}
		if child+1 < hi && data.Less(first+child, first+child+1) {
			child++
		}
		if !data.Less(first+root, first+child) {
			return
		}
		sliceSwap(data.x, first+root, first+child)
		root = child
	}
}

// Auto-generated variant of sort.gno:heapSort
func heapSort_func(data lessSwap, a, b int) {
	first := a
	lo := 0
	hi := b - a
	for i := (hi - 1) / 2; i >= 0; i-- {
		siftDown_func(data, i, hi, first)
	}
	for i := hi - 1; i >= 0; i-- {
		sliceSwap(data.x, first, first+i)
		siftDown_func(data, lo, i, first)
	}
}

// Auto-generated variant of sort.gno:medianOfThree
func medianOfThree_func(data lessSwap, m1, m0, m2 int) {
	if data.Less(m1, m0) {
		sliceSwap(data.x, m1, m0)
	}
	if data.Less(m2, m1) {
		sliceSwap(data.x, m2, m1)
		if data.Less(m1, m0) {
			sliceSwap(data.x, m1, m0)
		}
	}
}

// Auto-generated variant of sort.gno:swapRange
func swapRange_func(data lessSwap, a, b, n int) {
	for i := 0; i < n; i++ {
		sliceSwap(data.x, a+i, b+i)
	}
}

// Auto-generated variant of sort.gno:doPivot
func doPivot_func(data lessSwap, lo, hi int) (midlo, midhi int) {
	m := int(uint(lo+hi) >> 1)
	if hi-lo > 40 {
		s := (hi - lo) / 8
		medianOfThree_func(data, lo, lo+s, lo+2*s)
		medianOfThree_func(data, m, m-s, m+s)
		medianOfThree_func(data, hi-1, hi-1-s, hi-1-2*s)
	}
	medianOfThree_func(data, lo, m, hi-1)
	pivot := lo
	a, c := lo+1, hi-1
	for ; a < c && data.Less(a, pivot); a++ {
	}
	b := a
	for {
		for ; b < c && !data.Less(pivot, b); b++ {
		}
		for ; b < c && data.Less(pivot, c-1); c-- {
		}
		if b >= c {
			break
		}
		sliceSwap(data.x, b, c-1)
		b++
		c--
	}
	protect := hi-c < 5
	if !protect && hi-c < (hi-lo)/4 {
		dups := 0
		if !data.Less(pivot, hi-1) {
			sliceSwap(data.x, c, hi-1)
			c++
			dups++
		}
		if !data.Less(b-1, pivot) {
			b--
			dups++
		}
		if !data.Less(m, pivot) {
			sliceSwap(data.x, m, b-1)
			b--
			dups++
		}
		protect = dups > 1
	}
	if protect {
		for {
			for ; a < b && !data.Less(b-1, pivot); b-- {
			}
			for ; a < b && data.Less(a, pivot); a++ {
			}
			if a >= b {
				break
			}
			sliceSwap(data.x, a, b-1)
			a++
			b--
		}
	}
	sliceSwap(data.x, pivot, b-1)
	return b - 1, c
}

// Auto-generated variant of sort.gno:quickSort
func quickSort_func(data lessSwap, a, b, maxDepth int) {
	for b-a > 12 {
		if maxDepth == 0 {
			heapSort_func(data, a, b)
			return
		}
		maxDepth--
		mlo, mhi := doPivot_func(data, a, b)
		if mlo-a < b-mhi {
			quickSort_func(data, a, mlo, maxDepth)
			a = mhi
		} else {
			quickSort_func(data, mhi, b, maxDepth)
			b = mlo
		}
	}
	if b-a > 1 {
		for i := a + 6; i < b; i++ {
			if data.Less(i, i-6) {
				sliceSwap(data.x, i, i-6)
			}
		}
		insertionSort_func(data, a, b)
	}
}

// Auto-generated variant of sort.gno:stable
func stable_func(data lessSwap, n int) {
	blockSize := 20
	a, b := 0, blockSize
	for b <= n {
		insertionSort_func(data, a, b)
		a = b
		b += blockSize
	}
	insertionSort_func(data, a, n)
	for blockSize < n {
		a, b = 0, 2*blockSize
		for b <= n {
			symMerge_func(data, a, a+blockSize, b)
			a = b
			b += 2 * blockSize
		}
		if m := a + blockSize; m < n {
			symMerge_func(data, a, m, n)
		}
		blockSize *= 2
	}
}

// Auto-generated variant of sort.gno:symMerge
func symMerge_func(data lessSwap, a, m, b int) {
	if m-a == 1 {
		i := m
		j := b
		for i < j {
			h := int(uint(i+j) >> 1)
			if data.Less(h, a) {
				i = h + 1
			} else {
				j = h
			}
		}
		for k := a; k < i-1; k++ {
			sliceSwap(data.x, k, k+1)
		}
		return
	}
	if b-m == 1 {
		i := a
		j := m
		for i < j {
			h := int(uint(i+j) >> 1)
			if !data.Less(m, h) {
				i = h + 1
			} else {
				j = h
			}
		}
		for k := m; k > i; k-- {
			sliceSwap(data.x, k, k-1)
		}
		return
	}
	mid := int(uint(a+b) >> 1)
	n := mid + m
	var start, r int
	if m > mid {
		start = n - b
		r = mid
	} else {
		start = a
		r = m
	}
	p := n - 1
	for start < r {
		c := int(uint(start+r) >> 1)
		if !data.Less(p-c, c) {
			start = c + 1
		} else {
			r = c
		}
	}
	end := n - start
	if start < m && m < end {
		rotate_func(data, start, m, end)
	}
	if a < start && start < mid {
		symMerge_func(data, a, start, mid)
	}
	if mid < end && end < b {
		symMerge_func(data, mid, end, b)
	}
}

// Auto-generated variant of sort.gno:rotate
func rotate_func(data lessSwap, a, m, b int) {
	i := m - a
	j := b - m
	for i != j {
		if i > j {
			swapRange_func(data, m-i, m, j)
			i -= j
		} else {
			swapRange_func(data, m-i, m+j-i, i)
			j -= i
		}
	}
	swapRange_func(data, m-i, m, i)
}
