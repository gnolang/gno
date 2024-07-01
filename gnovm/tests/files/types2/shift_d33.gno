package main

import "fmt"

func main() {
	x := 11
	y := uint64((1 << 2) << x)
	println(y)
	fmt.Printf("%T \n", y)
}

// Output:
// 8192
// uint64
