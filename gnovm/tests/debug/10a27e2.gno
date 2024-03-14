package main

import "fmt"

func main() {
	x := 11
	//y := uint64(-(1 << 2) << x)
	y := uint64(-(1.0 << 2) << x)
	//y := uint64(-(1.2 << 2) << x)
	println(y)
	fmt.Printf("%T \n", y)
}

// Error:
// bigint underflows target kind
