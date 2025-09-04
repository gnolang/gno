package main

import "fmt"

func main() {
	x := 2
	var r float32
	r = 1<<x + 1 // untyped rhs to float32
	println(r)
	fmt.Printf("%T\n", r)
}

// Error:
// main/shift_c7.gno:8:2-14: operator << not defined on: Float32Kind

// TypeCheckError:
// main/shift_c7.gno:8:6: invalid operation: shifted operand 1 (type float32) must be integer
