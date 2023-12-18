package main

import "fmt"

type integer int

func main() { // TODO: refer to 47a, this shoul be solved
	// m(nil) is a conversion from nil to m, whose underlying type is map[string]int
	// lv: (nil main.m), rv: (undefined)
	if integer(nil) == nil {
		fmt.Println("integer is nil")
	} else {
		fmt.Println("integer is not nil")
	}
}

// Error:
// main/files/types/0f47_nilcast_filetest.gno:10: cannot convert (const (undefined) nil) to IntKind
