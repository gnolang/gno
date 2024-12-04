package main

import "fmt"

func main() {
	x := 2
	r := uint64(1 << x)
	println(r)
	fmt.Printf("%T\n", r)
}

// Output:
// 4
// uint64
