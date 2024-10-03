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

	var assertedErr interface{ IsNotSet() bool } // Define a variable of the interface type

	var ok bool

	// Perform the assertion and assign the result to assertedErr
	assertedErr, ok = err.(interface{ IsSet() bool }) // not impl lhs
	if ok {
		fmt.Println("Assertion succeeded:", ok)
		fmt.Println("IsSet:", assertedErr.IsSet())
	} else {
		fmt.Println("Assertion failed:", ok)
	}
}

// Error:
// main/files/types/assign_type_assertion_c.gno:23:2: interface{IsSet func()(bool)} does not implement interface{IsNotSet func()(bool)} (missing method IsNotSet)
