package main

import "fmt"

func main() {
	x := 11
	y := uint64(-(1.0 << x))
	println(y)
	fmt.Printf("%T \n", y)
}

// Error:
// main/files/debug/10a27c.gno:5: operator << not defined on: <untyped> bigdec
