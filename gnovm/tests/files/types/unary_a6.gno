package main

import "fmt"

func main() {
	x := 1.0
	y := ^x
	fmt.Println(y) // Output will be the bitwise complement of 1
}

// Error:
// main/files/types/unary_a6.gno:7: operator ^ not defined on: Float64Kind
