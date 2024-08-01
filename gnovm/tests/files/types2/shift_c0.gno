package main

import "fmt"

// case of const in type call
func main() {
	r := int(uint(1))
	println(r)
	fmt.Printf("%T \n", r)
}

// Output:
// 1
// int
