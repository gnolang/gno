package main

import "fmt"

func main() {
	x := 2
	r := uint64(1<<x + 1)
	println(r)
	fmt.Printf("%T \n", r)
}

// Output:
// 5
// uint64
