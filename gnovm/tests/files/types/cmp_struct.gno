package main

import "fmt"

// Define a struct that contains an array field
type Matrix struct {
	data [2]int
}

func main() {
	// Create two instances of the struct
	m1 := Matrix{data: [2]int{1, 2}}
	m2 := Matrix{data: [2]int{1, 2}}
	m3 := Matrix{data: [2]int{3, 4}}

	// Compare the instances
	fmt.Println("m1 == m2:", m1 == m2) // True because the data fields are identical
	fmt.Println("m1 == m3:", m1 == m3) // False because the data fields are different
}

// Output:
// m1 == m2: true
// m1 == m3: false
