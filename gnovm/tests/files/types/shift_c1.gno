package main

import "fmt"

// case of const in type call
func main() {
	r := int(uint(string("hello")))
	println(r)
	fmt.Printf("%T \n", r)
}

// Error:
// main/files/types/shift_c1.gno:7:11: cannot convert StringKind to UintKind
