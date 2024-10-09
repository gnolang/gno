package main

import "fmt"

func main() {
	x := 11
	y := uint64(-(1.0 << x)) // 1.0 is implicitly converted to uint64
	println(y)
	fmt.Printf("%T \n", y)
}

// Output:
// 18446744073709549568
// uint64
