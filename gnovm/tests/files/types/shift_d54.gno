package main

import "fmt"

func main() {
	a := 5 // infer type int
	var b int32 = 10
	c := b + a<<2

	fmt.Printf("%T, %d \n", c, c)
}

// Error:
// main/files/types/shift_d54.gno:8:7: invalid operation: mismatched types int32 and int
