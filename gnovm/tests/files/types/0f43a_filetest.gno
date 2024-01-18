package main

import "fmt"

type myPointer *int

func main() {
	if myPointer(nil) == nil {
		fmt.Println("Pointer is nil")
	} else {
		fmt.Println("Pointer is not nil")
	}
}

// Output:
// Pointer is nil
