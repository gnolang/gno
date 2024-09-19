package main

import "fmt"

func main() {
	var r uint64 = 1 << int8(1)
	fmt.Printf("%T \n", r)
	println(r)
}

// Output:
// uint64
// 2
