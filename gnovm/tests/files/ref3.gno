package main

import "fmt"

// ref of a concrete type; use fmt to print the pointer type.
// This verifies the type of the pointer itself, not just the deref.

func main() {
	var a int = 1
	b := &a
	fmt.Printf("%T\n", b)
}

// Output:
// *int
