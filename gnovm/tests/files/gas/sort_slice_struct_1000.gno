package main

import "sort"

// Sorts 1000 structs by Key with sort.Slice.
//
// sort_baseline_struct_1000.gno: same input, not sorted
// sort_interface_struct_1000.gno: same input, sorted through a sort.Interface

type item struct {
	Key  int
	ID   int
	Name string
}

func gen(n int) []item {
	s := make([]item, n)
	x := uint32(1)
	for i := range s {
		x = x*1664525 + 1013904223
		s[i] = item{Key: int(x >> 8), ID: i, Name: "item"}
	}
	return s
}

func main() {
	s := gen(1000)
	sort.Slice(s, func(i, j int) bool { return s[i].Key < s[j].Key })
	println(len(s))
}

// Output:
// 1000

// Gas:
// 57265367
