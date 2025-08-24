package main

import "fmt"

func main() {
	a := [2][2][]int{{{1, 2}, {3, 4}}, {{5, 6}, {7, 8}}}
	b := [2][2][]int{{{1, 2}, {3, 4}}, {{5, 6}, {7, 8}}}
	c := [2][2][]int{{{5, 6}, {7, 8}}, {{1, 2}, {3, 4}}}

	fmt.Println("a == b:", a == b)
	fmt.Println("a == c:", a == c)
}

// Error:
// main/cmp_array_c.gno:10:25-31: [2][]int is not comparable

// TypeCheckError:
// main/cmp_array_c.gno:10:25: invalid operation: a == b ([2][2][]int cannot be compared); main/cmp_array_c.gno:11:25: invalid operation: a == c ([2][2][]int cannot be compared)
