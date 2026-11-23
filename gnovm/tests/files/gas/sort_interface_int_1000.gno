package main

import "sort"

// Sorts 1000 ints through a user-written sort.Interface.
//
// sort_baseline_int_1000.gno: same input, not sorted

type byValue []int

func (s byValue) Len() int           { return len(s) }
func (s byValue) Less(i, j int) bool { return s[i] < s[j] }
func (s byValue) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

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
	sort.Sort(byValue(s))
	println(len(s))
}

// Output:
// 1000

// Gas:
// 67660600
