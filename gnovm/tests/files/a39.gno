package main

import "fmt"

func main() {
	a := [...]byte{}
	b := a
	fmt.Printf("%T %T\n", a, b)
}

// Go_Output:
// [0]uint8 [0]uint8

// Output:
// [0]uint8 [0]uint8
