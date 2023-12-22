package main

import "fmt"

func main() {
	x := 11
	y := 1 << x
	println(y)
	fmt.Printf("%T \n", y)
	fmt.Printf("%T \n", 1)
	fmt.Printf("%T \n", x)
}

// Output:
// 2048
// int
// int
// int
