package main

import "fmt"

func main() {
	var myPointer *int
	if myPointer == nil { // rcx.T == nil, <-,  after conversion lv: (nil *int), rv: (nil *int)
		fmt.Println("Pointer is nil")
	} else {
		fmt.Println("Pointer is not nil")
	}
}

// Output:
// Pointer is nil
