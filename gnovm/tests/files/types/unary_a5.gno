package main

import "fmt"

func main() {
	x := 1
	y := ^x
	fmt.Println(y) // Output will be the bitwise complement of 1
}

// Output:
// -2
