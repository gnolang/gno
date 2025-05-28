package main

import "fmt"

type integer int

func main() {
	// illegal conversion
	// should not work
	if integer(nil) == nil {
		fmt.Println("integer is nil")
	} else {
		fmt.Println("integer is not nil")
	}
}

// Error:
// main/typed_nil_a.gno:10:5-17: cannot convert (const (undefined)) to IntKind

// TypeCheckError:
// main/typed_nil_a.gno:10:13: cannot convert nil to type integer
