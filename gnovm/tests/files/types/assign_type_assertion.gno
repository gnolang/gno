package main

import (
	"fmt"
)

func main() {
	var i any = "Hello, world!"

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
// main/assign_type_assertion.gno:13:2-20: cannot use string as int

// TypeCheckError:
// main/assign_type_assertion.gno:13:10: cannot use i.(string) (comma, ok expression of type string) as int value in assignment
