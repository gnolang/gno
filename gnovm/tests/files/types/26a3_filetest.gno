package main

import "fmt"

func main() {
	a := uint(1)
	r := uint64(1) << a
	println(r)
	fmt.Printf("%T \n", r)

	//r1 := uint64(1) << (63 - 1)
	//println(r1)
}

// Output:
// 2
// uint64
