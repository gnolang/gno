package main

import "fmt"

func main() {
	a := [2][]int{{1, 2}, {3, 4}}
	b := [2][]int{{1, 2}, {3, 4}}
	c := [2][]int{{5, 6}, {7, 8}}

	fmt.Println("a == b:", a == b) // True because the elements match exactly
	fmt.Println("a == c:", a == c) // False because the elements do not match
}

// Error:
// main/cmp_array_b.gno:10:25-31: [2][]int is not comparable

// TypeCheckError:
// main/cmp_array_b.gno:10:25: invalid operation: a == b ([2][]int cannot be compared); main/cmp_array_b.gno:11:25: invalid operation: a == c ([2][]int cannot be compared)
