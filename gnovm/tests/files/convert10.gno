package main

import "fmt"

func main() {
	// Create a slice
	s := []int{1, 2, 3, 4, 5}

	// Convert slice to array pointer (length must match)
	var arrPtr *[5]int
	arrPtr = (*[5]int)(s) // Valid conversion

	fmt.Println("Original slice:", s)
	fmt.Println("Array pointer:", *arrPtr)

	// // Modify through the array pointer
	// arrPtr[0] = 100
	// fmt.Println("Modified slice:", s) // Underlying storage is shared!

	// // Convert back to slice (just for demonstration)
	// s2 := arrPtr[:]
	// fmt.Println("Converted back to slice:", s2)
}

// TODO: fix this

// Error:
// cannot convert []int to PointerKind
