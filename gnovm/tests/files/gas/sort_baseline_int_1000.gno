package main

import "sort"

// Generates the 1000-int input that sort_interface_int_1000.gno sorts, without
// sorting it. Subtract this file's gas to get the cost of the sort alone.

// Keep the sort import, so its load cost is in this file too.
var _ = sort.Sort

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
	println(len(s))
}

// Output:
// 1000

// Gas:
// 2637572
