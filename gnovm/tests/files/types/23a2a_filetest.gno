package main

import "fmt"

func main() {
	a := 1
	b := !a
	fmt.Println(b) // Output: false

}

// Error:
// main/files/types/23a2a_filetest.gno:7: operator ! not defined on: IntKind
