package main

import "fmt"

func main() {
	// A pointer to a pointer, initialized to nil
	var doublePtr **int = nil

	fmt.Println("1. Attempting to evaluate **doublePtr")

	// The expression is: *(*doublePtr)
	// The VM/Runtime must evaluate the INNER (*doublePtr) first
	// to get the address for the OUTER (*).
	
	// CRASH OCCURS HERE inside the first dereference.
	val := **doublePtr 

	// This line is never reached
	fmt.Println("2. Success:", val)
}