package main

import "fmt"

// case of expr in type call
func main() {
	x := 2
	var r float32
	r = 1 << x // untyped rhs to float32
	println(r)
	fmt.Printf("%T \n", r)
}

// Error:
// main/files/types2/shift_c6.gno:9:2: operator << not defined on: Float32Kind
