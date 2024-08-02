package main

import "fmt"

// case of expr in type call
func main() {
	x := 2
	var r float32
	r = 1<<x + 1 // untyped rhs to float32
	println(r)
	fmt.Printf("%T \n", r)
}

// Error:
// main/files/types/shift_c7.gno:9:2: operator << not defined on: Float32Kind
