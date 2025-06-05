package main

import "fmt"

func main() {
	var nilSlice []int // nil slice

	// Convert nil slice to [0]int (zero-length array)
	zeroArray := [0]int(nilSlice)

	fmt.Println("zeroArray: ", zeroArray)
}

// Output:
// zeroArray:  []
