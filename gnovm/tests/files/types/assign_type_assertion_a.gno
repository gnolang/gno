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

	var assertedErr interface{ IsSet() bool } // Define a variable of the interface type
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

// Output:
// Assertion succeeded: true
// IsSet: true
