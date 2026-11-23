package main

import "sort"

// Generates the 100-struct input that sort_interface_struct_100.gno sorts,
// without sorting it. Subtract this file's gas to get the cost of the sort
// alone.

type item struct {
	Key  int
	ID   int
	Name string
}

// Keep the sort import, so its load cost is in this file too.
var _ = sort.Sort

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
	s := gen(100)
	println(len(s))
}

// Output:
// 100

// Gas:
// 482659
