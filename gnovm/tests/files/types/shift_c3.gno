package main

import "fmt"

func main() {
	x := 1
	r := uint(+(1 << x))
	println(r)
	fmt.Printf("%T \n", r)
}

// Output:
// 2
// uint
