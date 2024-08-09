package main

import "fmt"

func main() {
	x := 1
	r := uint(+x)
	println(r)
	fmt.Printf("%T \n", r)
}

// Output:
// 1
// uint
