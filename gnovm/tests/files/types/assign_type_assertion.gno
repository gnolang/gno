package main

import (
	"fmt"
)

func main() {
	var i interface{} = "Hello, world!"

	// Attempt to assert the type of i to string
	var n int
	var ok bool
	n, ok = i.(string)
	if ok {
		fmt.Println("i contains a string:", n)
	} else {
		fmt.Println("i does not contain a string")
	}
}

// Error:
// main/files/types/assign_type_assertion.gno:13: cannot use string as int
