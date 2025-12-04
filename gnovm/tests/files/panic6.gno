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

// Stacktrace:
// panic: nil pointer dereference
// main<VPBlock(1,0)>()
//     main/panic6.gno:16

// Error:
// nil pointer dereference
