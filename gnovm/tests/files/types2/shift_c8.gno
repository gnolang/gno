package main

import "fmt"

// case of expr in type call
func main() {
	x := 2
	var r float32
	r = 1<<x + int(1) // typed rhs to float32
	println(r)
	fmt.Printf("%T \n", r)
}

// Error:
// main/files/types2/shift_c8.gno:9:2: cannot use int as float32
