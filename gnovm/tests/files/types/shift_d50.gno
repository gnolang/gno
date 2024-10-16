package main

import "fmt"

func main() {
	x := 11
	y := uint64(-(1.0 << 2) << x)
	println(y)
	fmt.Printf("%T \n", y)
}

// Error:
// main/files/types/shift_d50.gno:7:7: bigint underflows target kind
