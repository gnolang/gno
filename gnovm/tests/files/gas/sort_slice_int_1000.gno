package main

import "sort"

// Sorts 1000 ints with sort.Slice.
//
// sort_baseline_int_1000.gno: same input, not sorted
// sort_interface_int_1000.gno: same input, sorted through a sort.Interface

func gen(n int) []int {
	s := make([]int, n)
	x := uint32(1)
	for i := range s {
		x = x*1664525 + 1013904223
		s[i] = int(x >> 8)
	}
	return s
}

func main() {
	s := gen(1000)
	sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })
	println(len(s))
}

// Output:
// 1000

// Gas:
// 50673270
