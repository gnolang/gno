package main

import "fmt"

// Define a base type
type Base int

// Declare a new type that is a pointer to the base type
type PtrToBase *Base

func main() {
	var b Base = 42      // Initialize a variable of the base type
	var p PtrToBase = &b // Initialize a variable of the new pointer type with the address of b

	fmt.Printf("The value of b is: %d\n", b)

	// Using the new pointer type
	fmt.Printf("The value pointed to by p is: %d\n", *p)

	// Modifying the value pointed to by p
	*p = 100
	fmt.Printf("The new value of b after modification through p is: %d\n", b)
}

// Output:
// The value of b is: 42
// The value pointed to by p is: 42
// The new value of b after modification through p is: 100
