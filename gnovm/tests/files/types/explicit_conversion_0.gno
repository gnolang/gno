package main

import "fmt"

func main() {
	r := int(uint(1))
	println(r)
	fmt.Printf("%T \n", r)
}

// Output:
// 1
// int
