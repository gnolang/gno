package main

import "fmt"

func main() {
	println(any(nil) == (*int)(nil))
	fmt.Printf("%T\n", any(nil))
}

// Output:
// false
// <nil>
