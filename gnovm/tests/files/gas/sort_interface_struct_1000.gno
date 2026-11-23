package main

import "sort"

// Sorts 1000 structs by Key through a user-written sort.Interface.
//
// sort_baseline_struct_1000.gno: same input, not sorted

type item struct {
	Key  int
	ID   int
	Name string
}

type byKey []item

func (s byKey) Len() int           { return len(s) }
func (s byKey) Less(i, j int) bool { return s[i].Key < s[j].Key }
func (s byKey) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

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
	sort.Sort(byKey(s))
	println(len(s))
}

// Output:
// 1000

// Gas:
// 74253342
