package main

import "fmt"

func main() {
	x := 11
	y := 1.0 << x
	println(y)
	fmt.Printf("%T \n", y)
	fmt.Printf("%T \n", 1)
	fmt.Printf("%T \n", x)
}

// Error:
// main/files/types2/10a02_filetest.gno:5: operator << not defined on: <untyped> bigdec
