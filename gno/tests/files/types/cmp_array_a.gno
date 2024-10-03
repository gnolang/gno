package main

import "fmt"

// Define a struct that embeds an array
type Matrix struct {
	data [2]int
}

func main() {
	a := [2]Matrix{{data: [2]int{1, 2}}, {data: [2]int{3, 4}}}
	b := [2]Matrix{{data: [2]int{1, 2}}, {data: [2]int{3, 4}}}
	c := [2]Matrix{{data: [2]int{5, 6}}, {data: [2]int{7, 8}}}

	fmt.Println("a == b:", a == b) // True because the elements match exactly
	fmt.Println("a == c:", a == c) // False because the elements do not match
}

// Output:
// a == b: true
// a == c: false
