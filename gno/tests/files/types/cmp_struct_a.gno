package main

import "fmt"

// Define a struct that wraps the three-dimensional data structure
type Matrix struct {
	data [2][2][]int
}

func main() {
	a := Matrix{data: [2][2][]int{{{1, 2}, {3, 4}}, {{5, 6}, {7, 8}}}}
	b := Matrix{data: [2][2][]int{{{1, 2}, {3, 4}}, {{5, 6}, {7, 8}}}}
	c := Matrix{data: [2][2][]int{{{5, 6}, {7, 8}}, {{1, 2}, {3, 4}}}}

	fmt.Println("a == b:", a == b)
	fmt.Println("a == c:", a == c)
}

// Error:
// main/files/types/cmp_struct_a.gno:15:25: [2][]int is not comparable
