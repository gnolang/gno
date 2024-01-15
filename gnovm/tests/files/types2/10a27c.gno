package main

import "fmt"

func main() {
	x := 11
	y := uint64(-(1.0 << x))
	println(y)
	fmt.Printf("%T \n", y)
}

// Error:
// main/files/types2/10a27c.gno:7: operator << not defined on: BigdecKind
