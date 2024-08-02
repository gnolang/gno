package main

import "fmt"

func main() {
	x := 11
	y := uint64((+(x + 1) << 1) << 2)
	println(y)
	fmt.Printf("%T \n", y)
}

// Output:
// 96
// uint64
