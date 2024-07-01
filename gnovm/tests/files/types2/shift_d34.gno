package main

import "fmt"

func main() {
	x := 11
	y := uint64(-(1 << 2) << x)
	println(y)
	fmt.Printf("%T \n", y)
}

// Error:
// main/files/types2/shift_d34.gno:7: bigint underflows target kind
