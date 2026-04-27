package main

import "fmt"

// Deref of *interface{} when the interface holds nil.
// *p should be a nil interface{}, not a typed nil.

func main() {
	var i interface{}
	p := &i
	v := *p
	fmt.Println(v == nil)
}

// Output:
// true
