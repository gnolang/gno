package main

import "fmt"

func main() {
	a := map[any]int{[]int{1, 2}: 10, []int{2, 3}: 11, 3: 20}
	fmt.Println(a)
}

// Error:
// runtime error: slice type cannot be used as map key
