package main

import "fmt"

type integer int

func main() { // TODO: same with 47a
	// illegal conversion
	// should not work
	if integer(nil) == nil {
		fmt.Println("integer is nil")
	} else {
		fmt.Println("integer is not nil")
	}
}

// Error:
// main/files/types/0f47_nilcast_filetest.gno:10: cannot convert (const (undefined) nil) to IntKind
