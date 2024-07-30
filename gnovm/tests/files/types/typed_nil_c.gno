package main

import "fmt"

func main() {
	println(interface{}(nil) == (*int)(nil))
	fmt.Printf("%T \n", interface{}(nil))
}

// Output:
// false
// <nil>
