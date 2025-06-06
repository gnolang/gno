package main

import "fmt"

func main() {
	var nilSlice []int // nil slice

	// Convert nil slice to [0]int (zero-length array)
	zeroArray := [0]int(nilSlice)

	var zero2 [0]int

	fmt.Println("zeroArray: ", zeroArray)
	fmt.Println("zeroArray == zero2: ", zeroArray == zero2)
}

// Output:
// zeroArray:  []
// zeroArray == zero2:  true
