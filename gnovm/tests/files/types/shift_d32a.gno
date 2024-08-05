package main

import "fmt"

func main() {
	x := 11
	y := uint64(-(1.2 << x))
	println(y)
	fmt.Printf("%T \n", y)
}

// Error:
// main/files/types/shift_d32a.gno:7:7: cannot convert untyped bigdec to integer -- 1.2 not an exact integer
