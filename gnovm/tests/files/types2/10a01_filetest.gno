package main

import "fmt"

// TODO: this would work in Go.
// consider if we want.
func main() {
	x := 1
	y := 1.0 << 1
	println(y)
	fmt.Printf("%T \n", y)
	fmt.Printf("%T \n", 1)
	fmt.Printf("%T \n", x)

}

// Error:
// main/files/types2/10a01_filetest.gno:9: operator << not defined on: BigdecKind
