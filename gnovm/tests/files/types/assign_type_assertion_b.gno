package main

import "fmt"

type MyError struct{}

func (e MyError) Error() string {
	return "my error"
}

func (e MyError) IsSet() bool {
	return true
}

func main() {
	var err error = MyError{}

	var assertedErr int // Define a variable of the interface type
	var ok bool

	// Perform the assertion and assign the result to assertedErr
	assertedErr, ok = err.(interface{ IsSet() bool })
	if ok {
		fmt.Println("Assertion succeeded:", ok)
		fmt.Println("IsSet:", assertedErr.IsSet())
	} else {
		fmt.Println("Assertion failed:", ok)
	}
}

// Error:
// main/assign_type_assertion_b.gno:22:2-51: cannot use interface {IsSet func() bool} as int

// TypeCheckError:
// main/assign_type_assertion_b.gno:22:20: cannot use err.(interface{IsSet() bool}) (comma, ok expression of type interface{IsSet() bool}) as int value in assignment; main/assign_type_assertion_b.gno:25:37: assertedErr.IsSet undefined (type int has no field or method IsSet)
