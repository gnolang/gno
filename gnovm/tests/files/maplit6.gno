package main

import "fmt"

func main() {
	i1 := []int{1, 2}
	i2 := []int{2, 3}
	a := map[any]int{i1: 10, i2: 11, 3: 20}
	fmt.Println(a)
}

// Error:
// runtime error: slice type cannot be used as map key
