package main

import "fmt"

func main() {
	a := uint(1)
	r := uint64(1) << a
	println(r)
	fmt.Printf("%T\n", r)
}

// Output:
// 2
// uint64
