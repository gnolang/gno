package main

import "fmt"

func main() {
	x := 11
	y := uint64(-(1 << 2) << x)
	println(y)
	fmt.Printf("%T\n", y)
}

// Error:
// main/files/types/shift_d34.gno:7:7: bigint underflows target kind

// TypeCheckError:
// main/files/types/shift_d34.gno:7:14: -(1 << 2) (untyped int constant -4) overflows uint64
