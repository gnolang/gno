package main

import "fmt"

func main() {
	x := 11
	y := uint64(-(1.2 << x))
	println(y)
	fmt.Printf("%T\n", y)
}

// Error:
// main/shift_d32a.gno:7:16-24: invalid operation: shifted operand (const (1.2 <untyped> bigdec)) (<untyped> bigdec) must be integer

// TypeCheckError:
// main/shift_d32a.gno:7:16: invalid operation: shifted operand 1.2 (untyped float constant) must be integer
