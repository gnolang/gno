package main

import "fmt"

func main() {
	a := 1
	b := !a
	fmt.Println(b) // Output: false

}

// Error:
// main/files/types/unary_a2a.gno:7:7: operator ! not defined on: IntKind
