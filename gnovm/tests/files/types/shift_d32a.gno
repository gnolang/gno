package main

import "fmt"

func main() {
	x := 11
	y := uint64(-(1.2 << x))
	println(y)
	fmt.Printf("%T\n", y)
}

// Error:
// main/shift_d32a.gno:7:7-26: cannot convert untyped bigdec to integer -- 1.2 not an exact integer

// TypeCheckError:
// main/shift_d32a.gno:7:16: invalid operation: shifted operand 1.2 (untyped float constant) must be integer
