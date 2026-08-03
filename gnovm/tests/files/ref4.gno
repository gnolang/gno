package main

import "fmt"

// ref of a declared type; verify the pointer type string.

type myint int

func main() {
	var a myint = 1
	b := &a
	fmt.Printf("%T\n", b)
}

// Output:
// *main.myint
